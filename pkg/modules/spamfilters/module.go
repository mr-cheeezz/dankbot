package spamfilters

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/modules"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
)

const (
	floodWindow        = 10 * time.Second
	duplicateWindow    = time.Minute
	reloadInterval     = 10 * time.Second
	maxTimeoutDuration = 14 * 24 * time.Hour
)

var linkPattern = regexp.MustCompile(`(?i)(https?://|www\.)`)

type deleteMessageFunc func(context.Context, string) error
type timeoutUserFunc func(context.Context, string, time.Duration, string) error
type warnUserFunc func(context.Context, string, string) error
type banUserFunc func(context.Context, string, string) error

type messageSample struct {
	text string
	at   time.Time
}

type repeatOffenseState struct {
	count      int
	lastAt     time.Time
	memory     time.Duration
	untilReset bool
}

type Module struct {
	store *postgres.SpamFilterStore

	mu               sync.RWMutex
	filters          map[string]postgres.SpamFilter
	history          map[string][]messageSample
	repeatOffenses   map[string]repeatOffenseState
	lastOffensePrune time.Time
	deleteMessage    deleteMessageFunc
	timeoutUser      timeoutUserFunc
	warnUser         warnUserFunc
	banUser          banUserFunc
}

func New(store *postgres.SpamFilterStore) *Module {
	return &Module{
		store:          store,
		filters:        make(map[string]postgres.SpamFilter),
		history:        make(map[string][]messageSample),
		repeatOffenses: make(map[string]repeatOffenseState),
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

func (m *Module) SetModerationActions(
	deleteMessage deleteMessageFunc,
	timeoutUser timeoutUserFunc,
	warnUser warnUserFunc,
	banUser banUserFunc,
) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.deleteMessage = deleteMessage
	m.timeoutUser = timeoutUser
	m.warnUser = warnUser
	m.banUser = banUser
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
		return m.applyAction(ctx, *matched, "message length", now), nil
	}

	if matched := enabledFilter(filters, "links"); matched != nil && countLinks(message) > matched.ThresholdValue {
		return m.applyAction(ctx, *matched, "link filter", now), nil
	}

	if matched := enabledFilter(filters, "caps"); matched != nil && capsPercentage(message) >= matched.ThresholdValue {
		return m.applyAction(ctx, *matched, "caps filter", now), nil
	}

	if matched := enabledFilter(filters, "repeated-characters"); matched != nil && longestRepeatedRun(message) >= matched.ThresholdValue {
		return m.applyAction(ctx, *matched, "repeated characters filter", now), nil
	}

	samples := m.recordAndLoadHistory(ctx.SenderID, normalized, now)

	if matched := enabledFilter(filters, "message-flood"); matched != nil && recentCount(samples, floodWindow) >= matched.ThresholdValue {
		return m.applyAction(ctx, *matched, "message flood", now), nil
	}

	if matched := enabledFilter(filters, "duplicate-messages"); matched != nil && duplicateCount(samples, normalized, duplicateWindow) >= matched.ThresholdValue {
		return m.applyAction(ctx, *matched, "duplicate message", now), nil
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

func (m *Module) applyAction(ctx modules.CommandContext, filter postgres.SpamFilter, reason string, now time.Time) modules.MessageResult {
	action := strings.ToLower(strings.TrimSpace(filter.Action))
	deleteEnabled := strings.Contains(action, "delete")
	warnEnabled := strings.Contains(action, "warn")
	timeoutEnabled := strings.Contains(action, "timeout")
	banEnabled := strings.Contains(action, "ban")
	if !deleteEnabled && !warnEnabled && !timeoutEnabled && !banEnabled {
		// Keep legacy behavior for unknown action strings.
		deleteEnabled = true
	}

	m.mu.RLock()
	deleteMessage := m.deleteMessage
	timeoutUser := m.timeoutUser
	warnUser := m.warnUser
	banUser := m.banUser
	m.mu.RUnlock()

	if banEnabled && banUser != nil {
		if err := banUser(context.Background(), ctx.SenderID, reason); err != nil {
			fmt.Printf("spam filter ban error (%s): %v\n", filter.FilterKey, err)
		}
	}
	if timeoutEnabled && timeoutUser != nil && !banEnabled {
		duration := parseTimeoutDuration(action)
		if filter.RepeatOffendersEnabled {
			offenseCount := m.nextRepeatOffenseCount(ctx.SenderID, filter, now)
			duration = scaledTimeoutDuration(duration, filter.RepeatMultiplier, offenseCount)
		}
		if err := timeoutUser(context.Background(), ctx.SenderID, duration, reason); err != nil {
			fmt.Printf("spam filter timeout error (%s): %v\n", filter.FilterKey, err)
		}
	}
	if warnEnabled && warnUser != nil && !banEnabled && !timeoutEnabled {
		if err := warnUser(context.Background(), ctx.SenderID, reason); err != nil {
			fmt.Printf("spam filter warn error (%s): %v\n", filter.FilterKey, err)
		}
	}
	if deleteEnabled && strings.TrimSpace(ctx.MessageID) != "" && deleteMessage != nil {
		if err := deleteMessage(context.Background(), ctx.MessageID); err != nil {
			fmt.Printf("spam filter delete error (%s): %v\n", filter.FilterKey, err)
		}
	}

	return modules.MessageResult{StopProcessing: true}
}

func (m *Module) nextRepeatOffenseCount(senderID string, filter postgres.SpamFilter, now time.Time) int {
	senderID = strings.TrimSpace(senderID)
	filterKey := strings.TrimSpace(strings.ToLower(filter.FilterKey))
	if senderID == "" || filterKey == "" {
		return 1
	}

	memory := time.Duration(filter.RepeatMemorySeconds) * time.Second
	if memory <= 0 {
		memory = 10 * time.Minute
	}
	untilReset := filter.RepeatUntilStreamEnd
	stateKey := senderID + "|" + filterKey

	m.mu.Lock()
	defer m.mu.Unlock()

	state, ok := m.repeatOffenses[stateKey]
	if !ok || (!state.untilReset && now.Sub(state.lastAt) > state.memory) {
		state = repeatOffenseState{
			count:      1,
			lastAt:     now,
			memory:     memory,
			untilReset: untilReset,
		}
		m.repeatOffenses[stateKey] = state
		m.pruneRepeatOffensesLocked(now)
		return 1
	}

	state.count++
	state.lastAt = now
	state.memory = memory
	state.untilReset = untilReset
	m.repeatOffenses[stateKey] = state
	m.pruneRepeatOffensesLocked(now)
	return state.count
}

func (m *Module) pruneRepeatOffensesLocked(now time.Time) {
	if now.Sub(m.lastOffensePrune) < time.Minute {
		return
	}
	m.lastOffensePrune = now

	for key, state := range m.repeatOffenses {
		if state.untilReset {
			continue
		}
		if now.Sub(state.lastAt) > state.memory {
			delete(m.repeatOffenses, key)
		}
	}
}

func scaledTimeoutDuration(base time.Duration, multiplier float64, offenseCount int) time.Duration {
	if base <= 0 {
		base = 30 * time.Second
	}
	if multiplier < 1 || math.IsNaN(multiplier) || math.IsInf(multiplier, 0) {
		multiplier = 1
	}
	if offenseCount <= 1 {
		if base > maxTimeoutDuration {
			return maxTimeoutDuration
		}
		return base
	}

	scaledSeconds := base.Seconds() * math.Pow(multiplier, float64(offenseCount-1))
	if math.IsNaN(scaledSeconds) || math.IsInf(scaledSeconds, 1) {
		return maxTimeoutDuration
	}
	if scaledSeconds < 1 {
		scaledSeconds = 1
	}
	maxSeconds := maxTimeoutDuration.Seconds()
	if scaledSeconds > maxSeconds {
		scaledSeconds = maxSeconds
	}

	return time.Duration(scaledSeconds * float64(time.Second))
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
