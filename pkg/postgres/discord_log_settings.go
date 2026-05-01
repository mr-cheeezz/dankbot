package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type DiscordLogSettings struct {
	Enabled         bool
	ChannelID       string
	LogChatMessages bool
	LogModActions   bool
	LogAuditLogs    bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type DiscordLogSettingsStore struct {
	client *Client
}

func NewDiscordLogSettingsStore(client *Client) *DiscordLogSettingsStore {
	return &DiscordLogSettingsStore{client: client}
}

func DefaultDiscordLogSettings() DiscordLogSettings {
	return DiscordLogSettings{
		Enabled:         false,
		ChannelID:       "",
		LogChatMessages: false,
		LogModActions:   true,
		LogAuditLogs:    true,
	}
}

func (s *DiscordLogSettingsStore) EnsureDefault(ctx context.Context) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	defaults := DefaultDiscordLogSettings()
	_, err = db.ExecContext(
		ctx,
		`
INSERT INTO discord_log_settings (
	id,
	enabled,
	channel_id,
	log_chat_messages,
	log_mod_actions,
	log_audit_logs,
	created_at,
	updated_at
)
VALUES (1, $1, $2, $3, $4, $5, NOW(), NOW())
ON CONFLICT (id) DO NOTHING
`,
		defaults.Enabled,
		defaults.ChannelID,
		defaults.LogChatMessages,
		defaults.LogModActions,
		defaults.LogAuditLogs,
	)
	if err != nil {
		return fmt.Errorf("ensure discord log settings defaults: %w", err)
	}
	return nil
}

func (s *DiscordLogSettingsStore) Get(ctx context.Context) (*DiscordLogSettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var settings DiscordLogSettings
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	enabled,
	channel_id,
	log_chat_messages,
	log_mod_actions,
	log_audit_logs,
	created_at,
	updated_at
FROM discord_log_settings
WHERE id = 1
`,
	).Scan(
		&settings.Enabled,
		&settings.ChannelID,
		&settings.LogChatMessages,
		&settings.LogModActions,
		&settings.LogAuditLogs,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get discord log settings: %w", err)
	}

	settings.ChannelID = strings.TrimSpace(settings.ChannelID)
	return &settings, nil
}

func (s *DiscordLogSettingsStore) Update(ctx context.Context, settings DiscordLogSettings) (*DiscordLogSettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var updated DiscordLogSettings
	err = db.QueryRowContext(
		ctx,
		`
UPDATE discord_log_settings
SET
	enabled = $1,
	channel_id = $2,
	log_chat_messages = $3,
	log_mod_actions = $4,
	log_audit_logs = $5,
	updated_at = NOW()
WHERE id = 1
RETURNING
	enabled,
	channel_id,
	log_chat_messages,
	log_mod_actions,
	log_audit_logs,
	created_at,
	updated_at
`,
		settings.Enabled,
		strings.TrimSpace(settings.ChannelID),
		settings.LogChatMessages,
		settings.LogModActions,
		settings.LogAuditLogs,
	).Scan(
		&updated.Enabled,
		&updated.ChannelID,
		&updated.LogChatMessages,
		&updated.LogModActions,
		&updated.LogAuditLogs,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("update discord log settings: %w", err)
	}

	updated.ChannelID = strings.TrimSpace(updated.ChannelID)
	return &updated, nil
}
