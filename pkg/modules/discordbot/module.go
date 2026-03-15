package discordbot

import (
	"context"
	"fmt"
	"strings"

	"github.com/mr-cheeezz/dankbot/pkg/modules"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
)

type Module struct {
	settings   *postgres.DiscordBotSettingsStore
	guildID    string
	adminID    string
	sendToChan func(channelID, content string) error
}

func New(settings *postgres.DiscordBotSettingsStore, guildID, adminID string) *Module {
	return &Module{
		settings: settings,
		guildID:  strings.TrimSpace(guildID),
		adminID:  strings.TrimSpace(adminID),
	}
}

func (m *Module) Name() string {
	return "discord-bot"
}

func (m *Module) RegisterCommands() map[string]modules.CommandDefinition {
	return map[string]modules.CommandDefinition{
		"dping": {
			Handler:        m.pingRole,
			Description:    "Ping a configured Discord role in the linked server.",
			Usage:          "!dping <alias> [message]",
			Example:        "!dping announcements private server is live",
			CanDisable:     false,
			Configurable:   false,
			DefaultEnabled: true,
		},
	}
}

func (m *Module) Start(ctx context.Context) error {
	if m.settings == nil {
		return nil
	}
	return m.settings.EnsureDefault(ctx, m.guildID)
}

func (m *Module) SetDiscordSender(send func(channelID, content string) error) {
	m.sendToChan = send
}

func (m *Module) pingRole(ctx modules.CommandContext) (string, error) {
	if !m.canManageDiscord(ctx) {
		return "", nil
	}
	if m.settings == nil {
		return "Discord Bot settings are not configured yet.", nil
	}
	if m.sendToChan == nil {
		return "Discord Bot is not online right now.", nil
	}

	settings, err := m.settings.Get(context.Background())
	if err != nil {
		return "", err
	}
	if settings == nil {
		return "Discord Bot settings are not configured yet.", nil
	}
	if strings.TrimSpace(settings.DefaultChannelID) == "" {
		return "Set a Discord channel in the Discord Bot page first.", nil
	}
	if len(ctx.Args) == 0 {
		return "Use !dping <alias> [message].", nil
	}

	alias := normalizeAlias(ctx.Args[0])
	role := findPingRole(settings.PingRoles, alias)
	if role == nil || !role.Enabled {
		return fmt.Sprintf("No enabled Discord ping role is configured for %q.", alias), nil
	}

	message := strings.TrimSpace(strings.Join(ctx.Args[1:], " "))
	if len(message) > 400 {
		message = strings.TrimSpace(message[:400])
	}

	content := "<@&" + role.RoleID + ">"
	if message != "" {
		content += " " + message
	}

	if err := m.sendToChan(settings.DefaultChannelID, content); err != nil {
		return "", err
	}

	if strings.TrimSpace(role.RoleName) != "" {
		return "Pinged Discord role " + role.RoleName + ".", nil
	}
	return "Pinged Discord role " + alias + ".", nil
}

func (m *Module) canManageDiscord(ctx modules.CommandContext) bool {
	if ctx.IsBroadcaster || ctx.IsModerator {
		return true
	}
	return strings.TrimSpace(ctx.SenderID) != "" && strings.TrimSpace(ctx.SenderID) == m.adminID
}

func findPingRole(items []postgres.DiscordPingRole, alias string) *postgres.DiscordPingRole {
	for _, item := range items {
		if !item.Enabled {
			continue
		}
		if normalizeAlias(item.Alias) == alias {
			role := item
			return &role
		}
	}
	return nil
}

func normalizeAlias(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, "_", "-")
	fields := strings.FieldsFunc(value, func(r rune) bool {
		switch {
		case r >= 'a' && r <= 'z':
			return false
		case r >= '0' && r <= '9':
			return false
		case r == '-':
			return false
		default:
			return true
		}
	})
	return strings.Trim(strings.Join(fields, "-"), "-")
}
