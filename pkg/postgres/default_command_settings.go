package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type DefaultCommandSetting struct {
	CommandName string
	Enabled     bool
	ConfigJSON  json.RawMessage
	UpdatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type DefaultCommandSettingStore struct {
	client *Client
}

func NewDefaultCommandSettingStore(client *Client) *DefaultCommandSettingStore {
	return &DefaultCommandSettingStore{client: client}
}

func (s *DefaultCommandSettingStore) EnsureDefaults(ctx context.Context, settings []DefaultCommandSetting) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	for _, setting := range settings {
		setting.CommandName = strings.TrimSpace(strings.ToLower(setting.CommandName))
		if setting.CommandName == "" {
			continue
		}

		configJSON := normalizeConfigJSON(setting.ConfigJSON)

		if _, err := db.ExecContext(
			ctx,
			`
INSERT INTO default_command_settings (
	command_name,
	enabled,
	config_json,
	updated_by,
	created_at,
	updated_at
)
VALUES ($1, $2, $3, $4, NOW(), NOW())
ON CONFLICT (command_name) DO NOTHING
`,
			setting.CommandName,
			setting.Enabled,
			[]byte(configJSON),
			strings.TrimSpace(setting.UpdatedBy),
		); err != nil {
			return fmt.Errorf("ensure default command setting %q: %w", setting.CommandName, err)
		}
	}

	return nil
}

func (s *DefaultCommandSettingStore) Get(ctx context.Context, commandName string) (*DefaultCommandSetting, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	commandName = strings.TrimSpace(strings.ToLower(commandName))
	if commandName == "" {
		return nil, nil
	}

	var setting DefaultCommandSetting
	var configJSON []byte
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	command_name,
	enabled,
	config_json,
	updated_by,
	created_at,
	updated_at
FROM default_command_settings
WHERE command_name = $1
`,
		commandName,
	).Scan(
		&setting.CommandName,
		&setting.Enabled,
		&configJSON,
		&setting.UpdatedBy,
		&setting.CreatedAt,
		&setting.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get default command setting %q: %w", commandName, err)
	}

	setting.ConfigJSON = normalizeConfigJSON(configJSON)
	return &setting, nil
}

func (s *DefaultCommandSettingStore) List(ctx context.Context) ([]DefaultCommandSetting, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(
		ctx,
		`
SELECT
	command_name,
	enabled,
	config_json,
	updated_by,
	created_at,
	updated_at
FROM default_command_settings
ORDER BY command_name ASC
`,
	)
	if err != nil {
		return nil, fmt.Errorf("list default command settings: %w", err)
	}
	defer rows.Close()

	settings := make([]DefaultCommandSetting, 0)
	for rows.Next() {
		var setting DefaultCommandSetting
		var configJSON []byte
		if err := rows.Scan(
			&setting.CommandName,
			&setting.Enabled,
			&configJSON,
			&setting.UpdatedBy,
			&setting.CreatedAt,
			&setting.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan default command setting: %w", err)
		}
		setting.ConfigJSON = normalizeConfigJSON(configJSON)
		settings = append(settings, setting)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate default command settings: %w", err)
	}

	return settings, nil
}

func (s *DefaultCommandSettingStore) SetEnabled(ctx context.Context, commandName string, enabled bool, updatedBy string) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	commandName = strings.TrimSpace(strings.ToLower(commandName))
	if commandName == "" {
		return fmt.Errorf("command name is required")
	}

	result, err := db.ExecContext(
		ctx,
		`
UPDATE default_command_settings
SET
	enabled = $2,
	updated_by = $3,
	updated_at = NOW()
WHERE command_name = $1
`,
		commandName,
		enabled,
		strings.TrimSpace(updatedBy),
	)
	if err != nil {
		return fmt.Errorf("set default command enabled state %q: %w", commandName, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("set default command enabled state %q rows affected: %w", commandName, err)
	}
	if rows == 0 {
		return fmt.Errorf("default command setting %q does not exist", commandName)
	}

	return nil
}

func normalizeConfigJSON(raw []byte) json.RawMessage {
	raw = []byte(strings.TrimSpace(string(raw)))
	if len(raw) == 0 {
		return json.RawMessage(`{}`)
	}

	return json.RawMessage(raw)
}
