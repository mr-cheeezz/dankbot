package spamfilters

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/modules"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
)

const (
	floodWindow     = 10 * time.Second
	duplicateWindow = time.Minute
	reloadInterval  = 10 * time.Second
)

var linkPattern = regexp.MustCompile(`(?i)(https?://|www\.)`)

type deleteMessageFunc func(context.Context, string) error
type timeoutUserFunc func(context.Context, string, time.Duration, string) error

type messageSample struct {
	text string
	at   time.Time
}

type Module struct {
	store *postgres.SpamFilterStore

	mu            sync.RWMutex
	filters       map[string]postgres.SpamFilter
	history       map[string][]messageSample
	deleteMessage deleteMessageFunc
	timeoutUser   timeoutUserFunc
}

func New(store *postgres.SpamFilterStore) *Module {
	return &Module{
		store:   store,
		filters: make(map[string]postgres.SpamFilter),
		history: make(map[string][]messageSample),
	}
}

func (m *Module) Name() string {
	return "spam-filters"
}

func (m *Module) RegisterCommands() map[string]modules.CommandDefinition {
	return nil
}

func (m *Module) Start(ctx context.Context) error {
	if m.store == nil {
		return fmt.Errorf("spam filter store is not configured")
	}
	if err := m.store.EnsureDefaults(ctx); err != nil {
		return err
	}
	if err := m.reload(ctx); err != nil {
		return err
	}

	go m.runReloadLoop(ctx)
	return nil
}

func (m *Module) SetModerationActions(deleteMessage deleteMessageFunc, timeoutUser timeoutUserFunc) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.deleteMessage = deleteMessage
	m.timeoutUser = timeoutUser
}

func (m *Module) HandleMessage(ctx modules.CommandContext) (modules.MessageResult, error) {
	if ctx.Platform != "twitch" {
		return modules.MessageResult{}, nil
	}
	if ctx.IsBroadcaster || ctx.IsModerator {
		return modules.MessageResult{}, nil
	}

	message := strings.TrimSpace(ctx.Message)
	if message == "" {
		return modules.MessageResult{}, nil
	}

	now := time.Now().UTC()
	normalized := normalizeMessage(message)

	filters := m.snapshotFilters()
	if len(filters) == 0 {
		return modules.MessageResult{}, nil
	}

	if matched := enabledFilter(filters, "message-length"); matched != nil && len([]rune(message)) > matched.ThresholdValue {
		return m.applyAction(ctx, *matched, "message length"), nil
	}

	if matched := enabledFilter(filters, "links"); matched != nil && countLinks(message) > matched.ThresholdValue {
		return m.applyAction(ctx, *matched, "link filter"), nil
	}

	if matched := enabledFilter(filters, "caps"); matched != nil && capsPercentage(message) >= matched.ThresholdValue {
		return m.applyAction(ctx, *matched, "caps filter"), nil
	}

	if matched := enabledFilter(filters, "repeated-characters"); matched != nil && longestRepeatedRun(message) >= matched.ThresholdValue {
		return m.applyAction(ctx, *matched, "repeated characters filter"), nil
	}

	samples := m.recordAndLoadHistory(ctx.SenderID, normalized, now)

	if matched := enabledFilter(filters, "message-flood"); matched != nil && recentCount(samples, floodWindow) >= matched.ThresholdValue {
		return m.applyAction(ctx, *matched, "message flood"), nil
	}

	if matched := enabledFilter(filters, "duplicate-messages"); matched != nil && duplicateCount(samples, normalized, duplicateWindow) >= matched.ThresholdValue {
		return m.applyAction(ctx, *matched, "duplicate message"), nil
	}

	return modules.MessageResult{}, nil
}

func (m *Module) runReloadLoop(ctx context.Context) {
	ticker := time.NewTicker(reloadInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := m.reload(ctx); err != nil {
				fmt.Printf("spam filter reload error: %v\n", err)
			}
		}
	}
}

func (m *Module) reload(ctx context.Context) error {
	items, err := m.store.List(ctx)
	if err != nil {
		return err
	}

	next := make(map[string]postgres.SpamFilter, len(items))
	for _, item := range items {
		next[strings.TrimSpace(strings.ToLower(item.FilterKey))] = item
	}

	m.mu.Lock()
	m.filters = next
	m.mu.Unlock()

	return nil
}

func (m *Module) snapshotFilters() map[string]postgres.SpamFilter {
	m.mu.RLock()
	defer m.mu.RUnlock()

	next := make(map[string]postgres.SpamFilter, len(m.filters))
	for key, value := range m.filters {
		next[key] = value
	}

	return next
}

func (m *Module) recordAndLoadHistory(senderID, text string, now time.Time) []messageSample {
	senderID = strings.TrimSpace(senderID)
	if senderID == "" {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	current := m.history[senderID]
	trimmed := current[:0]
	cutoff := now.Add(-duplicateWindow)
	for _, sample := range current {
		if sample.at.After(cutoff) {
			trimmed = append(trimmed, sample)
		}
	}
	trimmed = append(trimmed, messageSample{text: text, at: now})
	m.history[senderID] = append([]messageSample(nil), trimmed...)

	return append([]messageSample(nil), trimmed...)
}

func (m *Module) applyAction(ctx modules.CommandContext, filter postgres.SpamFilter, reason string) modules.MessageResult {
	action := strings.ToLower(strings.TrimSpace(filter.Action))

	m.mu.RLock()
	deleteMessage := m.deleteMessage
	timeoutUser := m.timeoutUser
	m.mu.RUnlock()

	if strings.TrimSpace(ctx.MessageID) != "" && deleteMessage != nil {
		if err := deleteMessage(context.Background(), ctx.MessageID); err != nil {
			fmt.Printf("spam filter delete error (%s): %v\n", filter.FilterKey, err)
		}
	}

	if timeoutUser != nil && strings.Contains(action, "timeout") {
		duration := parseTimeoutDuration(action)
		if err := timeoutUser(context.Background(), ctx.SenderID, duration, reason); err != nil {
			fmt.Printf("spam filter timeout error (%s): %v\n", filter.FilterKey, err)
		}
	}

	return modules.MessageResult{StopProcessing: true}
}

func enabledFilter(filters map[string]postgres.SpamFilter, key string) *postgres.SpamFilter {
	filter, ok := filters[strings.ToLower(strings.TrimSpace(key))]
	if !ok || !filter.Enabled {
		return nil
	}
	return &filter
}

func normalizeMessage(message string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(message)), " "))
}

func countLinks(message string) int {
	return len(linkPattern.FindAllStringIndex(message, -1))
}

func capsPercentage(message string) int {
	letters := 0
	upper := 0
	for _, r := range message {
		if r >= 'A' && r <= 'Z' {
			letters++
			upper++
			continue
		}
		if r >= 'a' && r <= 'z' {
			letters++
		}
	}
	if letters < 8 {
		return 0
	}
	return int(float64(upper) / float64(letters) * 100)
}

func longestRepeatedRun(message string) int {
	maxRun := 0
	currentRun := 0
	var previous rune

	for _, r := range message {
		if r == previous {
			currentRun++
		} else {
			previous = r
			currentRun = 1
		}
		if currentRun > maxRun {
			maxRun = currentRun
		}
	}

	return maxRun
}

func recentCount(samples []messageSample, window time.Duration) int {
	if len(samples) == 0 {
		return 0
	}

	cutoff := time.Now().UTC().Add(-window)
	total := 0
	for _, sample := range samples {
		if sample.at.After(cutoff) {
			total++
		}
	}
	return total
}

func duplicateCount(samples []messageSample, text string, window time.Duration) int {
	if len(samples) == 0 || text == "" {
		return 0
	}

	cutoff := time.Now().UTC().Add(-window)
	total := 0
	for _, sample := range samples {
		if sample.text == text && sample.at.After(cutoff) {
			total++
		}
	}
	return total
}

func parseTimeoutDuration(action string) time.Duration {
	matches := regexp.MustCompile(`(\d+)\s*([smh]?)`).FindStringSubmatch(strings.ToLower(action))
	if len(matches) < 2 {
		return 30 * time.Second
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil || value <= 0 {
		return 30 * time.Second
	}

	switch matches[2] {
	case "h":
		return time.Duration(value) * time.Hour
	case "m":
		return time.Duration(value) * time.Minute
	default:
		return time.Duration(value) * time.Second
	}
}
