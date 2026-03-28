package defaultcommands

import (
	"context"
	"strings"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/modules"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
)

type Module struct {
	startedAt time.Time
	version   string
	settings  *postgres.DefaultCommandSettingStore
}

func New(
	startedAt time.Time,
	version string,
	settings *postgres.DefaultCommandSettingStore,
) *Module {
	return &Module{
		startedAt: startedAt,
		version:   strings.TrimSpace(version),
		settings:  settings,
	}
}

func (m *Module) Name() string {
	return "default-commands"
}

func (m *Module) RegisterCommands() map[string]modules.CommandDefinition {
	return map[string]modules.CommandDefinition{
		"ping": m.pingDefinition(),
	}
}

func (m *Module) Start(ctx context.Context) error {
	if m.settings == nil {
		return nil
	}

	defaults := make([]postgres.DefaultCommandSetting, 0, 3)
	for name, definition := range m.RegisterCommands() {
		defaults = append(defaults, postgres.DefaultCommandSetting{
			CommandName: name,
			Enabled:     definition.DefaultEnabled,
			ConfigJSON:  []byte(`{}`),
		})
	}

	return m.settings.EnsureDefaults(ctx, defaults)
}

func (m *Module) commandEnabled(ctx context.Context, name string, canDisable bool, defaultEnabled bool) (bool, error) {
	if !canDisable {
		return true, nil
	}
	if m.settings == nil {
		return defaultEnabled, nil
	}

	setting, err := m.settings.Get(ctx, name)
	if err != nil {
		return false, err
	}
	if setting == nil {
		return defaultEnabled, nil
	}

	return setting.Enabled, nil
}
