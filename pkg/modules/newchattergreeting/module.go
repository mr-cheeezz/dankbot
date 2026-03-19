package newchattergreeting

import (
	"context"
	"fmt"
	"hash/fnv"
	"strings"
	"sync"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/modules"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
)

const reloadInterval = 15 * time.Second

type Module struct {
	settingsStore *postgres.NewChatterGreetingModuleSettingsStore

	mu       sync.RWMutex
	settings postgres.NewChatterGreetingModuleSettings
	channel  string
	say      func(channel, message string) error
}

func New(settingsStore *postgres.NewChatterGreetingModuleSettingsStore) *Module {
	return &Module{
		settingsStore: settingsStore,
		settings:      postgres.DefaultNewChatterGreetingModuleSettings(),
	}
}

func (m *Module) Name() string {
	return "new-chatter-greeting"
}

func (m *Module) RegisterCommands() map[string]modules.CommandDefinition {
	return nil
}

func (m *Module) Start(ctx context.Context) error {
	if m.settingsStore == nil {
		return nil
	}
	if err := m.settingsStore.EnsureDefault(ctx); err != nil {
		return err
	}
	if err := m.reload(ctx); err != nil {
		return err
	}

	go m.runReloadLoop(ctx)
	return nil
}

func (m *Module) SetChatOutput(channel string, say func(channel, message string) error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.channel = strings.TrimSpace(channel)
	m.say = say
}

func (m *Module) HandleMessage(ctx modules.CommandContext) (modules.MessageResult, error) {
	if ctx.Platform != "twitch" || !ctx.FirstMessage {
		return modules.MessageResult{}, nil
	}
	if ctx.IsBroadcaster || ctx.IsModerator {
		return modules.MessageResult{}, nil
	}
	if hasCommandPrefix(ctx.Message, ctx.CommandPrefix) {
		return modules.MessageResult{}, nil
	}

	settings := m.snapshotSettings()
	if !settings.Enabled || len(settings.Messages) == 0 {
		return modules.MessageResult{}, nil
	}

	channel, say := m.output()
	if channel == "" || say == nil {
		return modules.MessageResult{}, nil
	}

	template := selectGreetingTemplate(settings.Messages, ctx.SenderID, ctx.DisplayName, ctx.Sender)
	message := renderGreetingTemplate(template, ctx)
	if strings.TrimSpace(message) == "" {
		return modules.MessageResult{}, nil
	}

	if err := say(channel, message); err != nil {
		fmt.Printf("new chatter greeting send error: %v\n", err)
	}

	return modules.MessageResult{}, nil
}

func hasCommandPrefix(message, prefix string) bool {
	message = strings.TrimSpace(message)
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		prefix = "!"
	}

	return strings.HasPrefix(message, prefix)
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
				fmt.Printf("new chatter greeting reload error: %v\n", err)
			}
		}
	}
}

func (m *Module) reload(ctx context.Context) error {
	settings, err := m.settingsStore.Get(ctx)
	if err != nil {
		return err
	}
	if settings == nil {
		defaults := postgres.DefaultNewChatterGreetingModuleSettings()
		settings = &defaults
	}

	m.mu.Lock()
	m.settings = *settings
	m.mu.Unlock()

	return nil
}

func (m *Module) snapshotSettings() postgres.NewChatterGreetingModuleSettings {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return postgres.NewChatterGreetingModuleSettings{
		Enabled:  m.settings.Enabled,
		Messages: append([]string(nil), m.settings.Messages...),
	}
}

func (m *Module) output() (string, func(channel, message string) error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.channel, m.say
}

func selectGreetingTemplate(messages []string, seedParts ...string) string {
	if len(messages) == 0 {
		return ""
	}
	if len(messages) == 1 {
		return messages[0]
	}

	h := fnv.New32a()
	for _, part := range seedParts {
		_, _ = h.Write([]byte(strings.TrimSpace(part)))
		_, _ = h.Write([]byte{':'})
	}
	index := int(h.Sum32() % uint32(len(messages)))
	return messages[index]
}

func renderGreetingTemplate(template string, ctx modules.CommandContext) string {
	template = strings.TrimSpace(template)
	if template == "" {
		return ""
	}

	login := strings.TrimSpace(ctx.Sender)
	displayName := strings.TrimSpace(ctx.DisplayName)
	if displayName == "" {
		displayName = login
	}
	if displayName == "" {
		displayName = "there"
	}

	user := "@" + displayName
	replacer := strings.NewReplacer(
		"{user}", user,
		"{target}", user,
		"{display_name}", displayName,
		"{login}", login,
		"{channel}", strings.TrimSpace(ctx.Channel),
	)

	return strings.TrimSpace(replacer.Replace(template))
}
