package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type QuoteModuleSettings struct {
	Enabled   bool
	UpdatedBy string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type QuoteModuleSettingsStore struct {
	client *Client
}

func NewQuoteModuleSettingsStore(client *Client) *QuoteModuleSettingsStore {
	return &QuoteModuleSettingsStore{client: client}
}

func DefaultQuoteModuleSettings() QuoteModuleSettings {
	return QuoteModuleSettings{
		Enabled: true,
	}
}

func (s *QuoteModuleSettingsStore) EnsureDefault(ctx context.Context) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	defaults := DefaultQuoteModuleSettings()
	_, err = db.ExecContext(
		ctx,
		`
INSERT INTO quote_module_settings (
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
		return fmt.Errorf("ensure quote module settings defaults: %w", err)
	}

	return nil
}

func (s *QuoteModuleSettingsStore) Get(ctx context.Context) (*QuoteModuleSettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var settings QuoteModuleSettings
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	enabled,
	updated_by,
	created_at,
	updated_at
FROM quote_module_settings
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
		return nil, fmt.Errorf("get quote module settings: %w", err)
	}

	return &settings, nil
}

func (s *QuoteModuleSettingsStore) Update(ctx context.Context, settings QuoteModuleSettings) (*QuoteModuleSettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var updated QuoteModuleSettings
	err = db.QueryRowContext(
		ctx,
		`
UPDATE quote_module_settings
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
		return nil, fmt.Errorf("update quote module settings: %w", err)
	}

	return &updated, nil
}
