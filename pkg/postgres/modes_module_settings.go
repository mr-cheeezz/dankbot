package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type ModesModuleSettings struct {
	LegacyCommandsEnabled bool
	UpdatedBy             string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type ModesModuleSettingsStore struct {
	client *Client
}

func NewModesModuleSettingsStore(client *Client) *ModesModuleSettingsStore {
	return &ModesModuleSettingsStore{client: client}
}

func DefaultModesModuleSettings() ModesModuleSettings {
	return ModesModuleSettings{
		LegacyCommandsEnabled: false,
	}
}

func (s *ModesModuleSettingsStore) EnsureDefault(ctx context.Context) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	defaults := DefaultModesModuleSettings()
	_, err = db.ExecContext(
		ctx,
		`
INSERT INTO modes_module_settings (
	id,
	legacy_commands_enabled,
	updated_by,
	created_at,
	updated_at
)
VALUES (1, $1, '', NOW(), NOW())
ON CONFLICT (id) DO NOTHING
`,
		defaults.LegacyCommandsEnabled,
	)
	if err != nil {
		return fmt.Errorf("ensure modes module settings defaults: %w", err)
	}

	return nil
}

func (s *ModesModuleSettingsStore) Get(ctx context.Context) (*ModesModuleSettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var settings ModesModuleSettings
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	legacy_commands_enabled,
	updated_by,
	created_at,
	updated_at
FROM modes_module_settings
WHERE id = 1
`,
	).Scan(
		&settings.LegacyCommandsEnabled,
		&settings.UpdatedBy,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get modes module settings: %w", err)
	}

	return &settings, nil
}

func (s *ModesModuleSettingsStore) Update(ctx context.Context, settings ModesModuleSettings) (*ModesModuleSettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var updated ModesModuleSettings
	err = db.QueryRowContext(
		ctx,
		`
UPDATE modes_module_settings
SET
	legacy_commands_enabled = $1,
	updated_by = $2,
	updated_at = NOW()
WHERE id = 1
RETURNING
	legacy_commands_enabled,
	updated_by,
	created_at,
	updated_at
`,
		settings.LegacyCommandsEnabled,
		strings.TrimSpace(settings.UpdatedBy),
	).Scan(
		&updated.LegacyCommandsEnabled,
		&updated.UpdatedBy,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("update modes module settings: %w", err)
	}

	return &updated, nil
}
