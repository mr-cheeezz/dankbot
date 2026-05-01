package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type RustLogModuleSettings struct {
	Enabled   bool
	UpdatedBy string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type RustLogModuleSettingsStore struct {
	client *Client
}

func NewRustLogModuleSettingsStore(client *Client) *RustLogModuleSettingsStore {
	return &RustLogModuleSettingsStore{client: client}
}

func DefaultRustLogModuleSettings(defaultEnabled bool) RustLogModuleSettings {
	return RustLogModuleSettings{
		Enabled: defaultEnabled,
	}
}

func (s *RustLogModuleSettingsStore) EnsureDefault(ctx context.Context, defaultEnabled bool) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	defaults := DefaultRustLogModuleSettings(defaultEnabled)
	_, err = db.ExecContext(
		ctx,
		`
INSERT INTO rustlog_module_settings (
	id,
	enabled,
	updated_by,
	created_at,
	updated_at
)
VALUES (1, $1, '', NOW(), NOW())
ON CONFLICT (id) DO NOTHING
`,
		defaults.Enabled,
	)
	if err != nil {
		return fmt.Errorf("ensure rustlog module settings defaults: %w", err)
	}

	return nil
}

func (s *RustLogModuleSettingsStore) Get(ctx context.Context) (*RustLogModuleSettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var settings RustLogModuleSettings
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	enabled,
	updated_by,
	created_at,
	updated_at
FROM rustlog_module_settings
WHERE id = 1
`,
	).Scan(
		&settings.Enabled,
		&settings.UpdatedBy,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get rustlog module settings: %w", err)
	}

	return &settings, nil
}

func (s *RustLogModuleSettingsStore) Update(ctx context.Context, settings RustLogModuleSettings) (*RustLogModuleSettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var updated RustLogModuleSettings
	err = db.QueryRowContext(
		ctx,
		`
UPDATE rustlog_module_settings
SET
	enabled = $1,
	updated_by = $2,
	updated_at = NOW()
WHERE id = 1
RETURNING
	enabled,
	updated_by,
	created_at,
	updated_at
`,
		settings.Enabled,
		strings.TrimSpace(settings.UpdatedBy),
	).Scan(
		&updated.Enabled,
		&updated.UpdatedBy,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("update rustlog module settings: %w", err)
	}

	return &updated, nil
}
