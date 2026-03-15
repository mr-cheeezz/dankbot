package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type FollowersOnlyModuleSettings struct {
	Enabled                 bool
	AutoDisableAfterMinutes int
	UpdatedBy               string
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

type FollowersOnlyModuleSettingsStore struct {
	client *Client
}

func NewFollowersOnlyModuleSettingsStore(client *Client) *FollowersOnlyModuleSettingsStore {
	return &FollowersOnlyModuleSettingsStore{client: client}
}

func DefaultFollowersOnlyModuleSettings() FollowersOnlyModuleSettings {
	return FollowersOnlyModuleSettings{
		Enabled:                 false,
		AutoDisableAfterMinutes: 30,
	}
}

func (s *FollowersOnlyModuleSettingsStore) EnsureDefault(ctx context.Context) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	defaults := DefaultFollowersOnlyModuleSettings()
	_, err = db.ExecContext(
		ctx,
		`
INSERT INTO followers_only_module_settings (
	id,
	enabled,
	auto_disable_after_minutes,
	updated_by,
	created_at,
	updated_at
)
VALUES (1, $1, $2, '', NOW(), NOW())
ON CONFLICT (id) DO NOTHING
`,
		defaults.Enabled,
		defaults.AutoDisableAfterMinutes,
	)
	if err != nil {
		return fmt.Errorf("ensure followers-only module settings defaults: %w", err)
	}

	return nil
}

func (s *FollowersOnlyModuleSettingsStore) Get(ctx context.Context) (*FollowersOnlyModuleSettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var settings FollowersOnlyModuleSettings
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	enabled,
	auto_disable_after_minutes,
	updated_by,
	created_at,
	updated_at
FROM followers_only_module_settings
WHERE id = 1
`,
	).Scan(
		&settings.Enabled,
		&settings.AutoDisableAfterMinutes,
		&settings.UpdatedBy,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get followers-only module settings: %w", err)
	}

	settings.AutoDisableAfterMinutes = normalizeFollowersOnlyAutoDisableMinutes(settings.AutoDisableAfterMinutes)
	return &settings, nil
}

func (s *FollowersOnlyModuleSettingsStore) Update(ctx context.Context, settings FollowersOnlyModuleSettings) (*FollowersOnlyModuleSettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var updated FollowersOnlyModuleSettings
	err = db.QueryRowContext(
		ctx,
		`
UPDATE followers_only_module_settings
SET
	enabled = $1,
	auto_disable_after_minutes = $2,
	updated_by = $3,
	updated_at = NOW()
WHERE id = 1
RETURNING
	enabled,
	auto_disable_after_minutes,
	updated_by,
	created_at,
	updated_at
`,
		settings.Enabled,
		normalizeFollowersOnlyAutoDisableMinutes(settings.AutoDisableAfterMinutes),
		strings.TrimSpace(settings.UpdatedBy),
	).Scan(
		&updated.Enabled,
		&updated.AutoDisableAfterMinutes,
		&updated.UpdatedBy,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("update followers-only module settings: %w", err)
	}

	updated.AutoDisableAfterMinutes = normalizeFollowersOnlyAutoDisableMinutes(updated.AutoDisableAfterMinutes)
	return &updated, nil
}

func normalizeFollowersOnlyAutoDisableMinutes(value int) int {
	switch {
	case value < 1:
		return 30
	case value > 24*60:
		return 24 * 60
	default:
		return value
	}
}
