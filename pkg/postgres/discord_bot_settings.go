package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type DiscordPingRole struct {
	Alias    string `json:"alias"`
	RoleID   string `json:"role_id"`
	RoleName string `json:"role_name"`
	Enabled  bool   `json:"enabled"`
}

type DiscordGamePingSettings struct {
	Enabled          bool     `json:"enabled"`
	ChannelID        string   `json:"channel_id"`
	RoleID           string   `json:"role_id"`
	RoleName         string   `json:"role_name"`
	MessageTemplate  string   `json:"message_template"`
	IncludeWatchLink bool     `json:"include_watch_link"`
	IncludeJoinLink  bool     `json:"include_join_link"`
	AllowedUsers     []string `json:"allowed_users"`
}

type DiscordBotSettings struct {
	GuildID          string
	DefaultChannelID string
	PingRoles        []DiscordPingRole
	GamePing         DiscordGamePingSettings
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type DiscordBotSettingsStore struct {
	client *Client
}

func NewDiscordBotSettingsStore(client *Client) *DiscordBotSettingsStore {
	return &DiscordBotSettingsStore{client: client}
}

func DefaultDiscordBotSettings(guildID string) DiscordBotSettings {
	return DiscordBotSettings{
		GuildID:          strings.TrimSpace(guildID),
		DefaultChannelID: "",
		PingRoles:        []DiscordPingRole{},
		GamePing:         defaultDiscordGamePingSettings(),
	}
}

func defaultDiscordGamePingSettings() DiscordGamePingSettings {
	return DiscordGamePingSettings{
		Enabled:          false,
		ChannelID:        "",
		RoleID:           "",
		RoleName:         "",
		MessageTemplate:  "NEW GAME: {game}",
		IncludeWatchLink: true,
		IncludeJoinLink:  true,
		AllowedUsers:     []string{},
	}
}

func (s *DiscordBotSettingsStore) EnsureDefault(ctx context.Context, guildID string) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	defaults := DefaultDiscordBotSettings(guildID)
	payload, err := json.Marshal(normalizeDiscordPingRoles(defaults.PingRoles))
	if err != nil {
		return fmt.Errorf("marshal discord bot settings defaults: %w", err)
	}
	gamePingPayload, err := json.Marshal(normalizeDiscordGamePingSettings(defaults.GamePing))
	if err != nil {
		return fmt.Errorf("marshal discord game ping defaults: %w", err)
	}

	_, err = db.ExecContext(
		ctx,
		`
INSERT INTO discord_bot_settings (
	id,
	guild_id,
	default_channel_id,
	ping_roles_json,
	game_ping_json,
	created_at,
	updated_at
)
VALUES (1, $1, $2, $3, $4, NOW(), NOW())
ON CONFLICT (id) DO NOTHING
`,
		strings.TrimSpace(defaults.GuildID),
		strings.TrimSpace(defaults.DefaultChannelID),
		payload,
		gamePingPayload,
	)
	if err != nil {
		return fmt.Errorf("ensure discord bot settings defaults: %w", err)
	}

	return nil
}

func (s *DiscordBotSettingsStore) Get(ctx context.Context) (*DiscordBotSettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var (
		settings DiscordBotSettings
		rawJSON  []byte
		rawGame  []byte
	)
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	guild_id,
	default_channel_id,
	ping_roles_json,
	game_ping_json,
	created_at,
	updated_at
FROM discord_bot_settings
WHERE id = 1
`,
	).Scan(
		&settings.GuildID,
		&settings.DefaultChannelID,
		&rawJSON,
		&rawGame,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get discord bot settings: %w", err)
	}

	if len(rawJSON) > 0 {
		if err := json.Unmarshal(rawJSON, &settings.PingRoles); err != nil {
			return nil, fmt.Errorf("decode discord bot ping roles: %w", err)
		}
	}
	if len(rawGame) > 0 {
		if err := json.Unmarshal(rawGame, &settings.GamePing); err != nil {
			return nil, fmt.Errorf("decode discord game ping settings: %w", err)
		}
	}

	settings.GuildID = strings.TrimSpace(settings.GuildID)
	settings.DefaultChannelID = strings.TrimSpace(settings.DefaultChannelID)
	settings.PingRoles = normalizeDiscordPingRoles(settings.PingRoles)
	settings.GamePing = normalizeDiscordGamePingSettings(settings.GamePing)

	return &settings, nil
}

func (s *DiscordBotSettingsStore) Update(ctx context.Context, settings DiscordBotSettings) (*DiscordBotSettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	normalized := DiscordBotSettings{
		GuildID:          strings.TrimSpace(settings.GuildID),
		DefaultChannelID: strings.TrimSpace(settings.DefaultChannelID),
		PingRoles:        normalizeDiscordPingRoles(settings.PingRoles),
		GamePing:         normalizeDiscordGamePingSettings(settings.GamePing),
	}

	payload, err := json.Marshal(normalized.PingRoles)
	if err != nil {
		return nil, fmt.Errorf("marshal discord bot ping roles: %w", err)
	}
	gamePingPayload, err := json.Marshal(normalized.GamePing)
	if err != nil {
		return nil, fmt.Errorf("marshal discord game ping settings: %w", err)
	}

	var (
		updated DiscordBotSettings
		rawJSON []byte
		rawGame []byte
	)
	err = db.QueryRowContext(
		ctx,
		`
UPDATE discord_bot_settings
SET
	guild_id = $1,
	default_channel_id = $2,
	ping_roles_json = $3,
	game_ping_json = $4,
	updated_at = NOW()
WHERE id = 1
RETURNING
	guild_id,
	default_channel_id,
	ping_roles_json,
	game_ping_json,
	created_at,
	updated_at
`,
		normalized.GuildID,
		normalized.DefaultChannelID,
		payload,
		gamePingPayload,
	).Scan(
		&updated.GuildID,
		&updated.DefaultChannelID,
		&rawJSON,
		&rawGame,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("update discord bot settings: %w", err)
	}

	if len(rawJSON) > 0 {
		if err := json.Unmarshal(rawJSON, &updated.PingRoles); err != nil {
			return nil, fmt.Errorf("decode updated discord bot ping roles: %w", err)
		}
	}
	if len(rawGame) > 0 {
		if err := json.Unmarshal(rawGame, &updated.GamePing); err != nil {
			return nil, fmt.Errorf("decode updated discord game ping settings: %w", err)
		}
	}
	updated.PingRoles = normalizeDiscordPingRoles(updated.PingRoles)
	updated.GamePing = normalizeDiscordGamePingSettings(updated.GamePing)

	return &updated, nil
}

func normalizeDiscordPingRoles(items []DiscordPingRole) []DiscordPingRole {
	if len(items) == 0 {
		return []DiscordPingRole{}
	}

	normalized := make([]DiscordPingRole, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		roleID := strings.TrimSpace(item.RoleID)
		alias := normalizeDiscordPingAlias(item.Alias)
		if roleID == "" || alias == "" {
			continue
		}
		if _, ok := seen[alias]; ok {
			continue
		}
		seen[alias] = struct{}{}
		normalized = append(normalized, DiscordPingRole{
			Alias:    alias,
			RoleID:   roleID,
			RoleName: strings.TrimSpace(item.RoleName),
			Enabled:  item.Enabled,
		})
	}

	return normalized
}

func normalizeDiscordPingAlias(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return ""
	}

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

func normalizeDiscordGamePingSettings(settings DiscordGamePingSettings) DiscordGamePingSettings {
	defaults := defaultDiscordGamePingSettings()
	looksLikeEmptyLegacy := !settings.Enabled &&
		!settings.IncludeWatchLink &&
		!settings.IncludeJoinLink &&
		strings.TrimSpace(settings.ChannelID) == "" &&
		strings.TrimSpace(settings.RoleID) == "" &&
		strings.TrimSpace(settings.MessageTemplate) == ""
	settings.ChannelID = strings.TrimSpace(settings.ChannelID)
	settings.RoleID = strings.TrimSpace(settings.RoleID)
	settings.RoleName = strings.TrimSpace(settings.RoleName)
	settings.MessageTemplate = strings.TrimSpace(settings.MessageTemplate)
	if settings.MessageTemplate == "" {
		settings.MessageTemplate = defaults.MessageTemplate
	}
	if looksLikeEmptyLegacy {
		settings.IncludeWatchLink = defaults.IncludeWatchLink
		settings.IncludeJoinLink = defaults.IncludeJoinLink
	}
	settings.AllowedUsers = normalizeDiscordAllowedUsers(settings.AllowedUsers)
	return settings
}

func normalizeDiscordAllowedUsers(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}

	normalized := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		value := strings.TrimSpace(strings.ToLower(item))
		value = strings.TrimPrefix(value, "@")
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	return normalized
}
