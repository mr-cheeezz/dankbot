package defaultcommands

import (
	"context"
	"fmt"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/modules"
)

func (m *Module) uptimeDefinition() modules.CommandDefinition {
	return modules.CommandDefinition{
		Handler:        m.uptime,
		Description:    "Shows how long the bot has been running.",
		Usage:          "!uptime",
		Example:        "!uptime",
		CanDisable:     true,
		Configurable:   false,
		DefaultEnabled: true,
	}
}

func (m *Module) uptime(ctx modules.CommandContext) (string, error) {
	_ = ctx

	enabled, err := m.commandEnabled(context.Background(), "uptime", true, true)
	if err != nil {
		return "", err
	}
	if !enabled {
		return "", nil
	}

	if m.startedAt.IsZero() {
		return "uptime unavailable", nil
	}

	return fmt.Sprintf("up for %s", time.Since(m.startedAt).Round(time.Second)), nil
}
