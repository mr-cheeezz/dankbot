package defaultcommands

import (
	"context"
	"fmt"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/modules"
)

func (m *Module) pingDefinition() modules.CommandDefinition {
	return modules.CommandDefinition{
		Handler:        m.ping,
		Description:    "Checks whether the bot is responding.",
		Usage:          "!ping",
		Example:        "!ping",
		CanDisable:     false,
		Configurable:   false,
		DefaultEnabled: true,
	}
}

func (m *Module) ping(ctx modules.CommandContext) (string, error) {
	_ = ctx

	enabled, err := m.commandEnabled(context.Background(), "ping", false, true)
	if err != nil {
		return "", err
	}
	if !enabled {
		return "", nil
	}

	version := m.version
	if version == "" {
		version = "dev"
	}

	if m.startedAt.IsZero() {
		return fmt.Sprintf("PONG. DankBot %s uptime is currently unavailable.", version), nil
	}

	return fmt.Sprintf(
		"PONG. DankBot %s has been up for %s since %s.",
		version,
		time.Since(m.startedAt).Round(time.Second),
		m.startedAt.UTC().Format("Jan 2, 2006 3:04 PM MST"),
	), nil
}
