package discordbot

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mr-cheeezz/dankbot/pkg/modules"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	robloxapi "github.com/mr-cheeezz/dankbot/pkg/roblox/api"
	steamapi "github.com/mr-cheeezz/dankbot/pkg/steam/api"
)

type Module struct {
	settings          *postgres.DiscordBotSettingsStore
	state             *postgres.BotStateStore
	robloxAccounts    *postgres.RobloxAccountStore
	robloxCookie      string
	guildID           string
	adminID           string
	sendToChan        func(channelID, content string) error
	sendEmbed         func(channelID, content string, embed *discordgo.MessageEmbed) error
	sendEmbedRich     func(channelID, content string, embed *discordgo.MessageEmbed, components []discordgo.MessageComponent) error
	isLive            func(context.Context) (bool, error)
	streamGameChecker func(context.Context) (bool, string, error)
	steamResolver     func(context.Context, string) (string, error)
}

func New(
	settings *postgres.DiscordBotSettingsStore,
	state *postgres.BotStateStore,
	robloxAccounts *postgres.RobloxAccountStore,
	robloxCookie string,
	guildID, adminID string,
) *Module {
	return &Module{
		settings:       settings,
		state:          state,
		robloxAccounts: robloxAccounts,
		robloxCookie:   strings.TrimSpace(robloxCookie),
		guildID:        strings.TrimSpace(guildID),
		adminID:        strings.TrimSpace(adminID),
	}
}

func (m *Module) Name() string {
	return "discord-bot"
}

func (m *Module) RegisterCommands() map[string]modules.CommandDefinition {
	return map[string]modules.CommandDefinition{
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

func (m *Module) SetDiscordRichEmbedSender(send func(channelID, content string, embed *discordgo.MessageEmbed, components []discordgo.MessageComponent) error) {
	m.sendEmbedRich = send
}

func (m *Module) SetStreamLiveChecker(checker func(context.Context) (bool, error)) {
	m.isLive = checker
}

func (m *Module) SetStreamGameChecker(checker func(context.Context) (bool, string, error)) {
	m.streamGameChecker = checker
}

func (m *Module) SetSteamResolver(resolver func(context.Context, string) (string, error)) {
	m.steamResolver = resolver
}

func (m *Module) canManageDiscord(ctx modules.CommandContext) bool {
	if ctx.IsBroadcaster || ctx.IsModerator {
		return true
	}
	return strings.TrimSpace(ctx.SenderID) != "" && strings.TrimSpace(ctx.SenderID) == m.adminID
}

func (m *Module) gamePing(ctx modules.CommandContext) (string, error) {
	if !m.canRunGamePing(ctx) {
		return "", nil
	}
	if m.settings == nil {
		return "Discord Bot settings are not configured yet.", nil
	}
	if m.sendEmbed == nil && m.sendEmbedRich == nil {
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
	if !m.isGamePingUserAllowed(ctx, settings.GamePing.AllowedUsers) {
		return "You are not allowed to run game pings.", nil
	}

	channelID := strings.TrimSpace(settings.GamePing.ChannelID)
	if channelID == "" {
		channelID = strings.TrimSpace(settings.DefaultChannelID)
	}
	if channelID == "" {
		return "Set a Discord channel in the Discord Bot page first.", nil
	}
	roleID := strings.TrimSpace(settings.GamePing.RoleID)
	roleName := strings.TrimSpace(settings.GamePing.RoleName)
	gameName := strings.TrimSpace(strings.Join(ctx.Args, " "))
	currentLive, currentGame, err := m.currentStreamState(context.Background())
	if err != nil {
		return "", err
	}
	if !currentLive {
		return "Game ping only works while the stream is live.", nil
	}
	if gameName == "" {
		gameName = currentGame
	}
	if strings.EqualFold(strings.TrimSpace(currentGame), "roblox") && (gameName == "" || strings.EqualFold(gameName, "roblox")) {
		if robloxExperience := m.currentRobloxExperienceName(context.Background()); robloxExperience != "" {
			gameName = robloxExperience
		}
	}
	if strings.TrimSpace(gameName) == "" {
		return "I could not determine the current stream game.", nil
	}
	if len(gameName) > 140 {
		gameName = strings.TrimSpace(gameName[:140])
	}

	description := strings.TrimSpace(settings.GamePing.MessageTemplate)
	if description == "" {
		description = "NEW GAME: {game}"
	}
	description = strings.ReplaceAll(description, "{game}", gameName)
	description = sanitizeDiscordMassMentions(description)

	watchURL := ""
	channelLogin := strings.TrimSpace(strings.TrimPrefix(ctx.Channel, "#"))
	if channelLogin != "" && settings.GamePing.IncludeWatchLink {
		watchURL = "https://twitch.tv/" + channelLogin
	}

	streamGameIsRoblox := strings.EqualFold(strings.TrimSpace(currentGame), "roblox")
	joinURL := ""
	joinLabel := ""

	if streamGameIsRoblox {
		if settings.GamePing.IncludeJoinLink {
			if link := m.robloxLinkModeURL(context.Background()); link != "" {
				joinURL = link
				joinLabel = "Join Roblox Server"
			} else {
				joinURL = robloxGameDiscoverURL(gameName)
				joinLabel = "Open Roblox Game"
			}
		}
	} else if settings.GamePing.IncludeJoinLink {
		storeURL := ""
		storeURL, err = m.resolveSteamURL(context.Background(), gameName)
		if err != nil {
			storeURL = ""
		}
		if strings.TrimSpace(storeURL) != "" {
			joinURL = strings.TrimSpace(storeURL)
			joinLabel = "Open on Steam"
		}
	}

	if watchURL != "" {
		description += "\n\n**Watch Live**\n" + watchURL
	}
	if joinURL != "" {
		description += "\n\n**Link**\n" + joinURL
	}

	content := ""
	if roleID != "" && canMentionGamePingRole(settings.GuildID, roleID, roleName) {
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

	components := buildGamePingLinkButtons(watchURL, joinURL, joinLabel)
	if m.sendEmbedRich != nil {
		if err := m.sendEmbedRich(channelID, content, embed, components); err != nil {
			return "", err
		}
		return "Sent Discord game ping for " + strconv.Quote(gameName) + ".", nil
	}

	if err := m.sendEmbed(channelID, content, embed); err != nil {
		return "", err
	}

	return "Sent Discord game ping for " + strconv.Quote(gameName) + ".", nil
}

func (m *Module) robloxLinkModeURL(ctx context.Context) string {
	if m.state == nil {
		return ""
	}

	state, err := m.state.Get(ctx)
	if err != nil || state == nil {
		return ""
	}
	if !strings.EqualFold(strings.TrimSpace(state.CurrentModeKey), "link") {
		return ""
	}

	link := strings.TrimSpace(state.CurrentModeParam)
	if strings.HasPrefix(strings.ToLower(link), "http://") || strings.HasPrefix(strings.ToLower(link), "https://") {
		return link
	}

	return ""
}

func robloxGameDiscoverURL(gameName string) string {
	gameName = strings.TrimSpace(gameName)
	if gameName == "" {
		return "https://www.roblox.com/discover"
	}

	return "https://www.roblox.com/discover/?Keyword=" + url.QueryEscape(gameName)
}

func (m *Module) canRunGamePing(ctx modules.CommandContext) bool {
	if m.canManageDiscord(ctx) {
		return true
	}
	return strings.TrimSpace(ctx.SenderID) != "" && strings.TrimSpace(ctx.SenderID) == m.adminID
}

func (m *Module) isGamePingUserAllowed(ctx modules.CommandContext, allowedUsers []string) bool {
	if m.canManageDiscord(ctx) {
		return true
	}

	allowed := normalizeAllowedUsers(allowedUsers)
	if len(allowed) == 0 {
		return false
	}

	senderLogin := normalizeAllowedUser(strings.TrimSpace(ctx.Sender))
	if senderLogin != "" {
		if _, ok := allowed[senderLogin]; ok {
			return true
		}
	}

	display := normalizeAllowedUser(strings.TrimSpace(ctx.DisplayName))
	if display != "" {
		if _, ok := allowed[display]; ok {
			return true
		}
	}

	return false
}

func (m *Module) currentStreamState(ctx context.Context) (bool, string, error) {
	if m.streamGameChecker != nil {
		return m.streamGameChecker(ctx)
	}
	if m.isLive != nil {
		live, err := m.isLive(ctx)
		return live, "", err
	}
	return false, "", nil
}

func (m *Module) resolveSteamURL(ctx context.Context, gameName string) (string, error) {
	gameName = strings.TrimSpace(gameName)
	if gameName == "" {
		return "", nil
	}
	if m.steamResolver != nil {
		return m.steamResolver(ctx, gameName)
	}

	client := steamapi.NewClient(nil, "")
	return client.ResolveStoreURL(ctx, gameName)
}

func (m *Module) currentRobloxExperienceName(ctx context.Context) string {
	if strings.TrimSpace(m.robloxCookie) == "" {
		return ""
	}

	client := robloxapi.NewClient(nil, m.robloxCookie)
	userID, err := m.robloxPresenceUserID(ctx, client)
	if err != nil || userID <= 0 {
		return ""
	}

	presences, err := client.GetPresences(ctx, []int64{userID})
	if err != nil || len(presences) == 0 {
		return ""
	}

	lastLocation := strings.TrimSpace(presences[0].LastLocation)
	if lastLocation == "" || strings.EqualFold(lastLocation, "website") {
		return ""
	}

	return lastLocation
}

func (m *Module) robloxPresenceUserID(ctx context.Context, client *robloxapi.Client) (int64, error) {
	if m.robloxAccounts != nil {
		account, err := m.robloxAccounts.Get(ctx, postgres.RobloxAccountKindStreamer)
		if err == nil && account != nil {
			linkedID, parseErr := strconv.ParseInt(strings.TrimSpace(account.RobloxUserID), 10, 64)
			if parseErr == nil && linkedID > 0 {
				return linkedID, nil
			}
		}
	}

	authUser, err := client.GetAuthenticatedUser(ctx)
	if err != nil {
		return 0, err
	}
	if authUser == nil || authUser.ID <= 0 {
		return 0, fmt.Errorf("roblox authenticated user id is missing")
	}

	return authUser.ID, nil
}

func normalizeAllowedUsers(users []string) map[string]struct{} {
	normalized := make(map[string]struct{}, len(users))
	for _, entry := range users {
		value := normalizeAllowedUser(entry)
		if value == "" {
			continue
		}
		normalized[value] = struct{}{}
	}
	return normalized
}

func normalizeAllowedUser(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.TrimPrefix(value, "@")
	return value
}

func buildGamePingLinkButtons(watchURL, joinURL, joinLabel string) []discordgo.MessageComponent {
	buttons := make([]discordgo.MessageComponent, 0, 2)

	if strings.TrimSpace(watchURL) != "" {
		buttons = append(buttons, discordgo.Button{
			Label: "Watch Live",
			Style: discordgo.LinkButton,
			URL:   strings.TrimSpace(watchURL),
		})
	}
	if strings.TrimSpace(joinURL) != "" {
		label := strings.TrimSpace(joinLabel)
		if label == "" {
			label = "Open Link"
		}
		buttons = append(buttons, discordgo.Button{
			Label: label,
			Style: discordgo.LinkButton,
			URL:   strings.TrimSpace(joinURL),
		})
	}

	if len(buttons) == 0 {
		return nil
	}

	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: buttons,
		},
	}
}

var timeNowUTC = func() time.Time {
	return time.Now().UTC()
}

func canMentionGamePingRole(guildID, roleID, roleName string) bool {
	roleID = strings.TrimSpace(roleID)
	if roleID == "" {
		return false
	}

	// Discord's @everyone role id equals the guild id.
	if strings.TrimSpace(guildID) != "" && strings.TrimSpace(guildID) == roleID {
		return false
	}

	name := strings.ToLower(strings.TrimSpace(roleName))
	if name == "@everyone" || name == "everyone" || name == "@here" || name == "here" {
		return false
	}

	return true
}

func sanitizeDiscordMassMentions(value string) string {
	value = strings.ReplaceAll(value, "@everyone", "@\u200beveryone")
	value = strings.ReplaceAll(value, "@here", "@\u200bhere")
	return value
}
