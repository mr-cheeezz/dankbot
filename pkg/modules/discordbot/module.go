package discordbot

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mr-cheeezz/dankbot/pkg/modules"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
)

type Module struct {
	settings   *postgres.DiscordBotSettingsStore
	state      *postgres.BotStateStore
	guildID    string
	adminID    string
	sendToChan func(channelID, content string) error
	sendEmbed  func(channelID, content string, embed *discordgo.MessageEmbed) error
	isLive     func(context.Context) (bool, error)
	liveMu     sync.Mutex
	lastLive   bool
	sentInLive bool
}

func New(settings *postgres.DiscordBotSettingsStore, state *postgres.BotStateStore, guildID, adminID string) *Module {
	return &Module{
		settings: settings,
		state:    state,
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
		"gameping": {
			Handler:        m.gamePing,
			Description:    "Sends a Discord game-change embed ping in the configured channel.",
			Usage:          "!gameping <game name>",
			Example:        "!gameping NFL Universe Football",
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

func (m *Module) SetDiscordEmbedSender(send func(channelID, content string, embed *discordgo.MessageEmbed) error) {
	m.sendEmbed = send
}

func (m *Module) SetStreamLiveChecker(checker func(context.Context) (bool, error)) {
	m.isLive = checker
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

func (m *Module) gamePing(ctx modules.CommandContext) (string, error) {
	if !m.canManageDiscord(ctx) {
		return "", nil
	}
	if m.settings == nil {
		return "Discord Bot settings are not configured yet.", nil
	}
	if m.sendEmbed == nil {
		return "Discord Bot is not online right now.", nil
	}

	settings, err := m.settings.Get(context.Background())
	if err != nil {
		return "", err
	}
	if settings == nil {
		return "Discord Bot settings are not configured yet.", nil
	}
	if !settings.GamePing.Enabled {
		return "Game ping is disabled in Discord Bot settings.", nil
	}

	channelID := strings.TrimSpace(settings.GamePing.ChannelID)
	if channelID == "" {
		channelID = strings.TrimSpace(settings.DefaultChannelID)
	}
	if channelID == "" {
		return "Set a Discord channel in the Discord Bot page first.", nil
	}
	args := append([]string(nil), ctx.Args...)
	roleID := strings.TrimSpace(settings.GamePing.RoleID)
	if len(args) > 0 {
		alias := normalizeAlias(args[0])
		if aliasRole := findPingRole(settings.PingRoles, alias); aliasRole != nil {
			roleID = strings.TrimSpace(aliasRole.RoleID)
			args = args[1:]
		}
	}

	gameName := strings.TrimSpace(strings.Join(args, " "))
	if gameName == "" {
		gameName = "game updated"
	}
	if len(gameName) > 140 {
		gameName = strings.TrimSpace(gameName[:140])
	}

	description := strings.TrimSpace(settings.GamePing.MessageTemplate)
	if description == "" {
		description = "NEW GAME: {game}"
	}
	description = strings.ReplaceAll(description, "{game}", gameName)

	watchURL := ""
	channelLogin := strings.TrimSpace(strings.TrimPrefix(ctx.Channel, "#"))
	if channelLogin != "" && settings.GamePing.IncludeWatchLink {
		watchURL = "https://twitch.tv/" + channelLogin
	}

	joinURL := ""
	if settings.GamePing.IncludeJoinLink && m.state != nil {
		state, stateErr := m.state.Get(context.Background())
		if stateErr == nil && state != nil &&
			strings.EqualFold(strings.TrimSpace(state.CurrentModeKey), "link") &&
			strings.HasPrefix(strings.TrimSpace(state.CurrentModeParam), "http") {
			joinURL = strings.TrimSpace(state.CurrentModeParam)
		}
	}

	if watchURL != "" {
		description += "\n\n**Watch Live**\n" + watchURL
	}
	if joinURL != "" {
		description += "\n\n**Join**\n" + joinURL
	}

	content := ""
	if m.allowGamePingRoleMention() && roleID != "" {
		content = "<@&" + roleID + ">"
	}

	footerName := strings.TrimSpace(ctx.DisplayName)
	if footerName == "" {
		footerName = strings.TrimSpace(ctx.Sender)
	}
	if footerName == "" {
		footerName = "dankbot"
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Game Change Ping",
		Description: description,
		Color:       0x5865F2,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Ping done by " + footerName,
		},
	}
	if watchURL != "" {
		embed.URL = watchURL
	}
	embed.Timestamp = timeNowUTC().Format(time.RFC3339)

	if err := m.sendEmbed(channelID, content, embed); err != nil {
		return "", err
	}

	return "Sent Discord game ping for " + strconv.Quote(gameName) + ".", nil
}

func (m *Module) allowGamePingRoleMention() bool {
	if m.isLive == nil {
		return true
	}
	live, err := m.isLive(context.Background())
	if err != nil {
		return true
	}

	m.liveMu.Lock()
	defer m.liveMu.Unlock()

	if !live {
		m.lastLive = false
		m.sentInLive = false
		return true
	}

	if !m.lastLive {
		m.lastLive = true
		m.sentInLive = true
		return false
	}
	if !m.sentInLive {
		m.sentInLive = true
		return false
	}
	return true
}

var timeNowUTC = func() time.Time {
	return time.Now().UTC()
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
