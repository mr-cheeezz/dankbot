package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type TabsModuleSettings struct {
	Enabled           bool
	InterestRatePct   float64
	InterestEveryDays int
	UpdatedBy         string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type TabsModuleSettingsStore struct {
	client *Client
}

func NewTabsModuleSettingsStore(client *Client) *TabsModuleSettingsStore {
	return &TabsModuleSettingsStore{client: client}
}

func DefaultTabsModuleSettings() TabsModuleSettings {
	return TabsModuleSettings{
		Enabled:           true,
		InterestRatePct:   0,
		InterestEveryDays: 7,
	}
}

func (s *TabsModuleSettingsStore) EnsureDefault(ctx context.Context) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	defaults := DefaultTabsModuleSettings()
	_, err = db.ExecContext(
		ctx,
		`
INSERT INTO tabs_module_settings (
	id,
	enabled,
	interest_rate_percent,
	interest_every_days,
	updated_by,
	created_at,
	updated_at
)
VALUES (1, $1, $2, $3, '', NOW(), NOW())
ON CONFLICT (id) DO NOTHING
`,
		defaults.Enabled,
		normalizeTabsInterestRate(defaults.InterestRatePct),
		normalizeTabsInterestEveryDays(defaults.InterestEveryDays),
	)
	if err != nil {
		return fmt.Errorf("ensure tabs module settings defaults: %w", err)
	}

	return nil
}

func (s *TabsModuleSettingsStore) Get(ctx context.Context) (*TabsModuleSettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var settings TabsModuleSettings
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	enabled,
	interest_rate_percent,
	interest_every_days,
	updated_by,
	created_at,
	updated_at
FROM tabs_module_settings
WHERE id = 1
`,
	).Scan(
		&settings.Enabled,
		&settings.InterestRatePct,
		&settings.InterestEveryDays,
		&settings.UpdatedBy,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get tabs module settings: %w", err)
	}

	settings.InterestRatePct = normalizeTabsInterestRate(settings.InterestRatePct)
	settings.InterestEveryDays = normalizeTabsInterestEveryDays(settings.InterestEveryDays)
	return &settings, nil
}

func (s *TabsModuleSettingsStore) Update(ctx context.Context, settings TabsModuleSettings) (*TabsModuleSettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var updated TabsModuleSettings
	err = db.QueryRowContext(
		ctx,
		`
UPDATE tabs_module_settings
SET
	enabled = $1,
	interest_rate_percent = $2,
	interest_every_days = $3,
	updated_by = $4,
	updated_at = NOW()
WHERE id = 1
RETURNING
	enabled,
	interest_rate_percent,
	interest_every_days,
	updated_by,
	created_at,
	updated_at
`,
		settings.Enabled,
		normalizeTabsInterestRate(settings.InterestRatePct),
		normalizeTabsInterestEveryDays(settings.InterestEveryDays),
		strings.TrimSpace(settings.UpdatedBy),
	).Scan(
		&updated.Enabled,
		&updated.InterestRatePct,
		&updated.InterestEveryDays,
		&updated.UpdatedBy,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("update tabs module settings: %w", err)
	}

	updated.InterestRatePct = normalizeTabsInterestRate(updated.InterestRatePct)
	updated.InterestEveryDays = normalizeTabsInterestEveryDays(updated.InterestEveryDays)
	return &updated, nil
}

func normalizeTabsInterestRate(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 500 {
		return 500
	}
	return value
}

func normalizeTabsInterestEveryDays(value int) int {
	switch {
	case value < 1:
		return 7
	case value > 365:
		return 365
	default:
		return value
	}
}
