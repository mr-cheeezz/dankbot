package dashboard

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
)

type discordBotChannelResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type discordBotRoleResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Mentionable bool   `json:"mentionable"`
}

type discordBotPingRoleResponse struct {
	Alias    string `json:"alias"`
	RoleID   string `json:"role_id"`
	RoleName string `json:"role_name"`
	Enabled  bool   `json:"enabled"`
}

type discordBotSettingsResponse struct {
	GuildID          string                       `json:"guild_id"`
	DefaultChannelID string                       `json:"default_channel_id"`
	PingRoles        []discordBotPingRoleResponse `json:"ping_roles"`
	GamePing         discordBotGamePingResponse   `json:"game_ping"`
	Logs             discordBotLogsResponse       `json:"logs"`
	Channels         []discordBotChannelResponse  `json:"channels"`
	Roles            []discordBotRoleResponse     `json:"roles"`
	CommandName      string                       `json:"command_name"`
	GamePingCommand  string                       `json:"game_ping_command_name"`
}

type discordBotGamePingResponse struct {
	Enabled          bool     `json:"enabled"`
	ChannelID        string   `json:"channel_id"`
	RoleID           string   `json:"role_id"`
	RoleName         string   `json:"role_name"`
	MessageTemplate  string   `json:"message_template"`
	IncludeWatchLink bool     `json:"include_watch_link"`
	IncludeJoinLink  bool     `json:"include_join_link"`
	AllowedUsers     []string `json:"allowed_users"`
}

type discordBotLogsResponse struct {
	Enabled         bool   `json:"enabled"`
	ChannelID       string `json:"channel_id"`
	LogChatMessages bool   `json:"log_chat_messages"`
	LogModActions   bool   `json:"log_mod_actions"`
	LogAuditLogs    bool   `json:"log_audit_logs"`
}

func (h handler) discordBot(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getDiscordBot(w, r)
	case http.MethodPut:
		h.updateDiscordBot(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet+", "+http.MethodPut)
	}
}

func (h handler) getDiscordBot(w http.ResponseWriter, r *http.Request) {
	if _, err := h.requireEditorFeatureAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	settings, logSettings, channels, roles, err := h.loadDiscordBotState(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(discordBotSettingsToResponse(settings, logSettings, channels, roles))
}

func (h handler) updateDiscordBot(w http.ResponseWriter, r *http.Request) {
	if _, err := h.requireEditorFeatureAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	settings, logSettings, channels, roles, err := h.loadDiscordBotState(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var request discordBotSettingsResponse
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid discord bot payload", http.StatusBadRequest)
		return
	}

	validChannels := make(map[string]struct{}, len(channels))
	for _, channel := range channels {
		validChannels[channel.ID] = struct{}{}
	}
	validRoles := make(map[string]string, len(roles))
	for _, role := range roles {
		validRoles[role.ID] = role.Name
	}

	defaultChannelID := strings.TrimSpace(request.DefaultChannelID)
	if defaultChannelID != "" {
		if _, ok := validChannels[defaultChannelID]; !ok {
			http.Error(w, "selected discord channel is not available in the connected guild", http.StatusBadRequest)
			return
		}
	}
	if request.Logs.ChannelID != "" {
		if _, ok := validChannels[strings.TrimSpace(request.Logs.ChannelID)]; !ok {
			http.Error(w, "selected logs channel is not available in the connected guild", http.StatusBadRequest)
			return
		}
	}

	nextPingRoles := make([]postgres.DiscordPingRole, 0, len(request.PingRoles))
	for _, item := range request.PingRoles {
		roleID := strings.TrimSpace(item.RoleID)
		roleName, ok := validRoles[roleID]
		if roleID == "" || !ok {
			continue
		}
		nextPingRoles = append(nextPingRoles, postgres.DiscordPingRole{
			Alias:    item.Alias,
			RoleID:   roleID,
			RoleName: roleName,
			Enabled:  item.Enabled,
		})
	}

	nextGamePing := postgres.DiscordGamePingSettings{
		Enabled:          request.GamePing.Enabled,
		ChannelID:        strings.TrimSpace(request.GamePing.ChannelID),
		RoleID:           strings.TrimSpace(request.GamePing.RoleID),
		MessageTemplate:  request.GamePing.MessageTemplate,
		IncludeWatchLink: request.GamePing.IncludeWatchLink,
		IncludeJoinLink:  request.GamePing.IncludeJoinLink,
		AllowedUsers:     request.GamePing.AllowedUsers,
	}
	if nextGamePing.ChannelID == "" {
		nextGamePing.ChannelID = defaultChannelID
	}
	if nextGamePing.ChannelID != "" {
		if _, ok := validChannels[nextGamePing.ChannelID]; !ok {
			http.Error(w, "selected game ping channel is not available in the connected guild", http.StatusBadRequest)
			return
		}
	}
	if nextGamePing.RoleID != "" {
		roleName, ok := validRoles[nextGamePing.RoleID]
		if !ok {
			http.Error(w, "selected game ping role is not available in the connected guild", http.StatusBadRequest)
			return
		}
		nextGamePing.RoleName = roleName
	}
	nextLogs := postgres.DiscordLogSettings{
		Enabled:         request.Logs.Enabled,
		ChannelID:       strings.TrimSpace(request.Logs.ChannelID),
		LogChatMessages: request.Logs.LogChatMessages,
		LogModActions:   request.Logs.LogModActions,
		LogAuditLogs:    request.Logs.LogAuditLogs,
	}
	if nextLogs.ChannelID == "" {
		nextLogs.ChannelID = strings.TrimSpace(logSettings.ChannelID)
	}

	updated, err := h.appState.DiscordBotSettings.Update(r.Context(), postgres.DiscordBotSettings{
		GuildID:          settings.GuildID,
		DefaultChannelID: defaultChannelID,
		PingRoles:        nextPingRoles,
		GamePing:         nextGamePing,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if updated == nil {
		http.Error(w, "discord bot settings not found", http.StatusNotFound)
		return
	}
	if h.appState != nil && h.appState.DiscordLogSettings != nil {
		if err := h.appState.DiscordLogSettings.EnsureDefault(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		updatedLogs, err := h.appState.DiscordLogSettings.Update(r.Context(), nextLogs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if updatedLogs != nil {
			logSettings = updatedLogs
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(discordBotSettingsToResponse(updated, logSettings, channels, roles))
}

func (h handler) loadDiscordBotState(r *http.Request) (*postgres.DiscordBotSettings, *postgres.DiscordLogSettings, []discordBotChannelResponse, []discordBotRoleResponse, error) {
	if h.appState == nil || h.appState.Config == nil || h.appState.DiscordBotInstallation == nil || h.appState.DiscordBotSettings == nil {
		return nil, nil, nil, nil, errors.New("discord bot settings are not configured")
	}

	token := strings.TrimSpace(h.appState.Config.Discord.BotToken)
	if token == "" || strings.EqualFold(token, "your_discord_bot_token") {
		return nil, nil, nil, nil, errors.New("discord bot token is not configured")
	}

	installation, err := h.appState.DiscordBotInstallation.Get(r.Context())
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if installation == nil || strings.TrimSpace(installation.GuildID) == "" {
		return nil, nil, nil, nil, errors.New("discord bot is not installed in a server yet")
	}

	if err := h.appState.DiscordBotSettings.EnsureDefault(r.Context(), installation.GuildID); err != nil {
		return nil, nil, nil, nil, err
	}

	settings, err := h.appState.DiscordBotSettings.Get(r.Context())
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if settings == nil {
		defaults := postgres.DefaultDiscordBotSettings(installation.GuildID)
		settings = &defaults
	}
	if strings.TrimSpace(settings.GuildID) == "" {
		settings.GuildID = strings.TrimSpace(installation.GuildID)
	}
	logSettings := postgres.DefaultDiscordLogSettings()
	if h.appState.DiscordLogSettings != nil {
		if err := h.appState.DiscordLogSettings.EnsureDefault(r.Context()); err != nil {
			return nil, nil, nil, nil, err
		}
		loadedLogs, err := h.appState.DiscordLogSettings.Get(r.Context())
		if err != nil {
			return nil, nil, nil, nil, err
		}
		if loadedLogs != nil {
			logSettings = *loadedLogs
		}
	}

	channels, roles, err := fetchDiscordGuildResources(token, settings.GuildID)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return settings, &logSettings, channels, roles, nil
}

func fetchDiscordGuildResources(token, guildID string) ([]discordBotChannelResponse, []discordBotRoleResponse, error) {
	session, err := discordgo.New("Bot " + strings.TrimSpace(token))
	if err != nil {
		return nil, nil, err
	}

	rawChannels, err := session.GuildChannels(strings.TrimSpace(guildID))
	if err != nil {
		return nil, nil, err
	}
	rawRoles, err := session.GuildRoles(strings.TrimSpace(guildID))
	if err != nil {
		return nil, nil, err
	}

	channels := make([]discordBotChannelResponse, 0, len(rawChannels))
	for _, channel := range rawChannels {
		if channel == nil {
			continue
		}
		if channel.Type != discordgo.ChannelTypeGuildText && channel.Type != discordgo.ChannelTypeGuildNews {
			continue
		}
		channels = append(channels, discordBotChannelResponse{
			ID:   strings.TrimSpace(channel.ID),
			Name: "#" + strings.TrimSpace(channel.Name),
		})
	}
	sort.Slice(channels, func(i, j int) bool {
		return strings.ToLower(channels[i].Name) < strings.ToLower(channels[j].Name)
	})

	roles := make([]discordBotRoleResponse, 0, len(rawRoles))
	for _, role := range rawRoles {
		if role == nil {
			continue
		}
		if role.Managed || role.Name == "@everyone" {
			continue
		}
		roles = append(roles, discordBotRoleResponse{
			ID:          strings.TrimSpace(role.ID),
			Name:        strings.TrimSpace(role.Name),
			Mentionable: role.Mentionable,
		})
	}
	sort.Slice(roles, func(i, j int) bool {
		return strings.ToLower(roles[i].Name) < strings.ToLower(roles[j].Name)
	})

	return channels, roles, nil
}

func discordBotSettingsToResponse(
	settings *postgres.DiscordBotSettings,
	logSettings *postgres.DiscordLogSettings,
	channels []discordBotChannelResponse,
	roles []discordBotRoleResponse,
) discordBotSettingsResponse {
	response := discordBotSettingsResponse{
		Channels:        channels,
		Roles:           roles,
		CommandName:     "!dping",
		GamePingCommand: "!gameping",
	}

	if settings != nil {
		response.GuildID = settings.GuildID
		response.DefaultChannelID = settings.DefaultChannelID
		response.PingRoles = make([]discordBotPingRoleResponse, 0, len(settings.PingRoles))
		for _, item := range settings.PingRoles {
			response.PingRoles = append(response.PingRoles, discordBotPingRoleResponse{
				Alias:    item.Alias,
				RoleID:   item.RoleID,
				RoleName: item.RoleName,
				Enabled:  item.Enabled,
			})
		}
		response.GamePing = discordBotGamePingResponse{
			Enabled:          settings.GamePing.Enabled,
			ChannelID:        settings.GamePing.ChannelID,
			RoleID:           settings.GamePing.RoleID,
			RoleName:         settings.GamePing.RoleName,
			MessageTemplate:  settings.GamePing.MessageTemplate,
			IncludeWatchLink: settings.GamePing.IncludeWatchLink,
			IncludeJoinLink:  settings.GamePing.IncludeJoinLink,
			AllowedUsers:     settings.GamePing.AllowedUsers,
		}
	}
	if logSettings != nil {
		response.Logs = discordBotLogsResponse{
			Enabled:         logSettings.Enabled,
			ChannelID:       strings.TrimSpace(logSettings.ChannelID),
			LogChatMessages: logSettings.LogChatMessages,
			LogModActions:   logSettings.LogModActions,
			LogAuditLogs:    logSettings.LogAuditLogs,
		}
	}

	return response
}
