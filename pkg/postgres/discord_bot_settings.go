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

type DiscordBotSettings struct {
	GuildID          string
	DefaultChannelID string
	PingRoles        []DiscordPingRole
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

	_, err = db.ExecContext(
		ctx,
		`
INSERT INTO discord_bot_settings (
	id,
	guild_id,
	default_channel_id,
	ping_roles_json,
	created_at,
	updated_at
)
VALUES (1, $1, $2, $3, NOW(), NOW())
ON CONFLICT (id) DO NOTHING
`,
		strings.TrimSpace(defaults.GuildID),
		strings.TrimSpace(defaults.DefaultChannelID),
		payload,
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
	)
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	guild_id,
	default_channel_id,
	ping_roles_json,
	created_at,
	updated_at
FROM discord_bot_settings
WHERE id = 1
`,
	).Scan(
		&settings.GuildID,
		&settings.DefaultChannelID,
		&rawJSON,
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

	settings.GuildID = strings.TrimSpace(settings.GuildID)
	settings.DefaultChannelID = strings.TrimSpace(settings.DefaultChannelID)
	settings.PingRoles = normalizeDiscordPingRoles(settings.PingRoles)

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
	}

	payload, err := json.Marshal(normalized.PingRoles)
	if err != nil {
		return nil, fmt.Errorf("marshal discord bot ping roles: %w", err)
	}

	var (
		updated DiscordBotSettings
		rawJSON []byte
	)
	err = db.QueryRowContext(
		ctx,
		`
UPDATE discord_bot_settings
SET
	guild_id = $1,
	default_channel_id = $2,
	ping_roles_json = $3,
	updated_at = NOW()
WHERE id = 1
RETURNING
	guild_id,
	default_channel_id,
	ping_roles_json,
	created_at,
	updated_at
`,
		normalized.GuildID,
		normalized.DefaultChannelID,
		payload,
	).Scan(
		&updated.GuildID,
		&updated.DefaultChannelID,
		&rawJSON,
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
	updated.PingRoles = normalizeDiscordPingRoles(updated.PingRoles)

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
