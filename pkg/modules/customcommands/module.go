package customcommands

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/modules"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
)

const refreshInterval = 5 * time.Second

type Module struct {
	store             *postgres.CustomCommandStore
	streamLiveChecker func() bool

	mu      sync.RWMutex
	entries []postgres.CustomCommand
}

func New(store *postgres.CustomCommandStore) *Module {
	return &Module{store: store}
}

func (m *Module) Name() string {
	return "custom-commands"
}

func (m *Module) RegisterCommands() map[string]modules.CommandDefinition {
	return nil
}

func (m *Module) SetStreamLiveChecker(checker func() bool) {
	m.streamLiveChecker = checker
}

func (m *Module) Start(ctx context.Context) error {
	if m.store == nil {
		return nil
	}
	if err := m.reload(ctx); err != nil {
		return err
	}
	go m.refreshLoop(ctx)
	return nil
}

func (m *Module) HandleMessage(ctx modules.CommandContext) (modules.MessageResult, error) {
	prefix := strings.TrimSpace(ctx.CommandPrefix)
	if prefix == "" || !strings.HasPrefix(strings.TrimSpace(ctx.Message), prefix) {
		return modules.MessageResult{}, nil
	}

	commandLine := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(ctx.Message), prefix))
	if commandLine == "" {
		return modules.MessageResult{}, nil
	}
	parts := strings.Fields(commandLine)
	if len(parts) == 0 {
		return modules.MessageResult{}, nil
	}

	candidates := []string{strings.ToLower(parts[0])}
	if len(parts) > 1 {
		candidates = append(candidates, strings.ToLower(parts[0]+" "+parts[1]))
	}

	entry := m.match(ctx.Platform, candidates...)
	if entry == nil {
		return modules.MessageResult{}, nil
	}
	if !entry.Enabled {
		return modules.MessageResult{StopProcessing: true}, nil
	}

	isLive := true
	if m.streamLiveChecker != nil {
		isLive = m.streamLiveChecker()
	}
	if isLive && !entry.EnabledWhenOnline {
		return modules.MessageResult{StopProcessing: true}, nil
	}
	if !isLive && !entry.EnabledWhenOffline {
		return modules.MessageResult{StopProcessing: true}, nil
	}

	reply := renderTemplate(entry.ResponseTemplate, ctx)
	if reply == "" {
		return modules.MessageResult{StopProcessing: true}, nil
	}
	if entry.ResponseType == "action" && ctx.Platform == "twitch" {
		reply = "/me " + reply
	}

	return modules.MessageResult{
		Reply:          reply,
		StopProcessing: true,
	}, nil
}

func (m *Module) refreshLoop(ctx context.Context) {
	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = m.reload(ctx)
		}
	}
}

func (m *Module) reload(ctx context.Context) error {
	if m.store == nil {
		return nil
	}
	entries, err := m.store.List(ctx)
	if err != nil {
		return err
	}
	m.mu.Lock()
	m.entries = entries
	m.mu.Unlock()
	return nil
}

func (m *Module) match(platform string, candidates ...string) *postgres.CustomCommand {
	normalizedPlatform := strings.ToLower(strings.TrimSpace(platform))
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, candidate := range candidates {
		normalizedCandidate := strings.ToLower(strings.TrimSpace(candidate))
		if normalizedCandidate == "" {
			continue
		}
		for _, entry := range m.entries {
			if entry.Platform != normalizedPlatform {
				continue
			}
			if entry.Name == normalizedCandidate {
				next := entry
				return &next
			}
			for _, alias := range entry.Aliases {
				if alias == normalizedCandidate {
					next := entry
					return &next
				}
			}
		}
	}

	return nil
}

func renderTemplate(template string, ctx modules.CommandContext) string {
	message := strings.TrimSpace(template)
	if message == "" {
		return ""
	}

	displayName := strings.TrimSpace(ctx.DisplayName)
	if displayName == "" {
		displayName = strings.TrimSpace(ctx.Sender)
	}

	replacements := map[string]string{
		"{user}":         displayName,
		"{sender}":       displayName,
		"{display_name}": displayName,
		"{channel}":      strings.TrimPrefix(strings.TrimSpace(ctx.Channel), "#"),
		"{args}":         strings.TrimSpace(strings.Join(ctx.Args, " ")),
	}
	for key, value := range replacements {
		message = strings.ReplaceAll(message, key, value)
	}
	return strings.TrimSpace(message)
}
