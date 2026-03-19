package blockedterms

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/modules"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
)

const reloadInterval = 10 * time.Second

type deleteMessageFunc func(context.Context, string) error
type timeoutUserFunc func(context.Context, string, time.Duration, string) error
type warnUserFunc func(context.Context, string, string) error
type banUserFunc func(context.Context, string, string) error

type compiledTerm struct {
	postgres.BlockedTerm
	regex                  *regexp.Regexp
	normalizedPhraseGroups [][]string
}

type Module struct {
	store *postgres.BlockedTermStore

	mu            sync.RWMutex
	terms         []compiledTerm
	deleteMessage deleteMessageFunc
	timeoutUser   timeoutUserFunc
	warnUser      warnUserFunc
	banUser       banUserFunc
}

func New(store *postgres.BlockedTermStore) *Module {
	return &Module{store: store}
}

func (m *Module) Name() string {
	return "blocked-terms"
}

func (m *Module) RegisterCommands() map[string]modules.CommandDefinition {
	return nil
}

func (m *Module) Start(ctx context.Context) error {
	if m.store == nil {
		return fmt.Errorf("blocked term store is not configured")
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

	term := m.matchingTerm(message)
	if term == nil {
		return modules.MessageResult{}, nil
	}

	m.applyAction(ctx, *term)
	return modules.MessageResult{StopProcessing: true}, nil
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
				fmt.Printf("blocked terms reload error: %v\n", err)
			}
		}
	}
}

func (m *Module) reload(ctx context.Context) error {
	items, err := m.store.List(ctx)
	if err != nil {
		return err
	}

	next := make([]compiledTerm, 0, len(items))
	for _, item := range items {
		if !item.Enabled {
			continue
		}

		compiled := compiledTerm{BlockedTerm: item}
		if item.IsRegex {
			regex, compileErr := regexp.Compile(item.Pattern)
			if compileErr != nil {
				fmt.Printf("blocked term regex compile error (%s): %v\n", item.ID, compileErr)
				continue
			}
			compiled.regex = regex
		} else {
			compiled.normalizedPhraseGroups = normalizePhraseGroups(item.PhraseGroups)
		}

		next = append(next, compiled)
	}

	m.mu.Lock()
	m.terms = next
	m.mu.Unlock()

	return nil
}

func (m *Module) matchingTerm(message string) *compiledTerm {
	m.mu.RLock()
	defer m.mu.RUnlock()

	normalized := strings.ToLower(message)
	for _, term := range m.terms {
		if term.IsRegex {
			if term.regex != nil && term.regex.MatchString(message) {
				copyTerm := term
				return &copyTerm
			}
			continue
		}

		if len(term.normalizedPhraseGroups) > 0 {
			for _, group := range term.normalizedPhraseGroups {
				if phraseGroupMatches(group, normalized) {
					copyTerm := term
					return &copyTerm
				}
			}
			continue
		}

		if strings.Contains(normalized, strings.ToLower(term.Pattern)) {
			copyTerm := term
			return &copyTerm
		}
	}

	return nil
}

func (m *Module) applyAction(ctx modules.CommandContext, term compiledTerm) {
	m.mu.RLock()
	deleteMessage := m.deleteMessage
	timeoutUser := m.timeoutUser
	warnUser := m.warnUser
	banUser := m.banUser
	m.mu.RUnlock()

	action := strings.TrimSpace(strings.ToLower(term.Action))
	reason := strings.TrimSpace(term.Reason)
	if reason == "" {
		reason = "Blocked term detected."
	}

	reason += " - Automated by dankbot"

	if strings.Contains(action, "delete") && strings.TrimSpace(ctx.MessageID) != "" && deleteMessage != nil {
		if err := deleteMessage(context.Background(), ctx.MessageID); err != nil {
			fmt.Printf("blocked term delete error (%s): %v\n", term.ID, err)
		}
	}

	switch action {
	case "warn", "delete + warn":
		if warnUser != nil {
			if err := warnUser(context.Background(), ctx.SenderID, reason); err != nil {
				fmt.Printf("blocked term warn error (%s): %v\n", term.ID, err)
			}
		}
	case "timeout", "delete + timeout":
		if timeoutUser != nil {
			duration := time.Duration(term.TimeoutSeconds) * time.Second
			if duration <= 0 {
				duration = 10 * time.Minute
			}
			if err := timeoutUser(context.Background(), ctx.SenderID, duration, reason); err != nil {
				fmt.Printf("blocked term timeout error (%s): %v\n", term.ID, err)
			}
		}
	case "ban", "delete + ban":
		if banUser != nil {
			if err := banUser(context.Background(), ctx.SenderID, reason); err != nil {
				fmt.Printf("blocked term ban error (%s): %v\n", term.ID, err)
			}
		}
	}
}

func normalizePhraseGroups(groups [][]string) [][]string {
	normalized := make([][]string, 0, len(groups))
	for _, group := range groups {
		nextGroup := make([]string, 0, len(group))
		for _, phrase := range group {
			value := strings.TrimSpace(strings.ToLower(phrase))
			if value == "" {
				continue
			}
			nextGroup = append(nextGroup, value)
		}
		if len(nextGroup) == 0 {
			continue
		}
		normalized = append(normalized, nextGroup)
	}

	return normalized
}

func phraseGroupMatches(group []string, normalizedMessage string) bool {
	if len(group) == 0 {
		return false
	}

	for _, phrase := range group {
		if !strings.Contains(normalizedMessage, phrase) {
			return false
		}
	}

	return true
}
