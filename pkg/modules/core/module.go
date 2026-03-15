package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/modules"
)

type Module struct {
	startedAt     time.Time
	commandSource interface {
		Names() []string
	}
}

func New(startedAt time.Time, commandSource interface{ Names() []string }) *Module {
	return &Module{
		startedAt:     startedAt,
		commandSource: commandSource,
	}
}

func (m *Module) Name() string {
	return "core"
}

func (m *Module) RegisterCommands() map[string]modules.CommandDefinition {
	return map[string]modules.CommandDefinition{
		"ping": {
			Handler:     m.ping,
			Description: "Checks whether the bot is responding.",
			Usage:       "!ping",
			Example:     "!ping",
		},
		"uptime": {
			Handler:     m.uptime,
			Description: "Shows how long the bot has been running.",
			Usage:       "!uptime",
			Example:     "!uptime",
		},
		"help": {
			Handler:     m.help,
			Description: "Lists available built-in commands.",
			Usage:       "!help",
			Example:     "!help",
		},
	}
}

func (m *Module) Start(ctx context.Context) error {
	_ = ctx
	return nil
}

func (m *Module) ping(ctx modules.CommandContext) (string, error) {
	_ = ctx
	return "pong", nil
}

func (m *Module) uptime(ctx modules.CommandContext) (string, error) {
	_ = ctx

	if m.startedAt.IsZero() {
		return "uptime unavailable", nil
	}

	return fmt.Sprintf("up for %s", time.Since(m.startedAt).Round(time.Second)), nil
}

func (m *Module) help(ctx modules.CommandContext) (string, error) {
	_ = ctx

	if m.commandSource == nil {
		return "no commands available", nil
	}

	names := m.commandSource.Names()
	if len(names) == 0 {
		return "no commands available", nil
	}

	return "commands: !" + strings.Join(names, ", !"), nil
}
