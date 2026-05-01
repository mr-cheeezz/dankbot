package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type TabsModuleSettings struct {
	Enabled                    bool
	InterestRatePct            float64
	InterestIntervalMode       string
	InterestIntervalCustomDays int
	GracePeriodDays            int
	UpdatedBy                  string
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
}

type TabsModuleSettingsStore struct {
	client *Client
}

func NewTabsModuleSettingsStore(client *Client) *TabsModuleSettingsStore {
	return &TabsModuleSettingsStore{client: client}
}

func DefaultTabsModuleSettings() TabsModuleSettings {
	return TabsModuleSettings{
		Enabled:                    true,
		InterestRatePct:            0,
		InterestIntervalMode:       "weekly",
		InterestIntervalCustomDays: 7,
		GracePeriodDays:            7,
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
	interest_interval_mode,
	interest_interval_custom_days,
	grace_period_days,
	updated_by,
	created_at,
	updated_at
)
VALUES (1, $1, $2, $3, $4, $5, '', NOW(), NOW())
ON CONFLICT (id) DO NOTHING
`,
		defaults.Enabled,
		normalizeTabsInterestRate(defaults.InterestRatePct),
		normalizeTabsInterestIntervalMode(defaults.InterestIntervalMode),
		normalizeTabsInterestIntervalCustomDays(defaults.InterestIntervalCustomDays),
		normalizeTabsGracePeriodDays(defaults.GracePeriodDays),
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
	interest_interval_mode,
	interest_interval_custom_days,
	grace_period_days,
	updated_by,
	created_at,
	updated_at
FROM tabs_module_settings
WHERE id = 1
`,
	).Scan(
		&settings.Enabled,
		&settings.InterestRatePct,
		&settings.InterestIntervalMode,
		&settings.InterestIntervalCustomDays,
		&settings.GracePeriodDays,
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
	settings.InterestIntervalMode = normalizeTabsInterestIntervalMode(settings.InterestIntervalMode)
	settings.InterestIntervalCustomDays = normalizeTabsInterestIntervalCustomDays(settings.InterestIntervalCustomDays)
	settings.GracePeriodDays = normalizeTabsGracePeriodDays(settings.GracePeriodDays)
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
	interest_interval_mode = $3,
	interest_interval_custom_days = $4,
	grace_period_days = $5,
	updated_by = $6,
	updated_at = NOW()
WHERE id = 1
RETURNING
	enabled,
	interest_rate_percent,
	interest_interval_mode,
	interest_interval_custom_days,
	grace_period_days,
	updated_by,
	created_at,
	updated_at
`,
		settings.Enabled,
		normalizeTabsInterestRate(settings.InterestRatePct),
		normalizeTabsInterestIntervalMode(settings.InterestIntervalMode),
		normalizeTabsInterestIntervalCustomDays(settings.InterestIntervalCustomDays),
		normalizeTabsGracePeriodDays(settings.GracePeriodDays),
		strings.TrimSpace(settings.UpdatedBy),
	).Scan(
		&updated.Enabled,
		&updated.InterestRatePct,
		&updated.InterestIntervalMode,
		&updated.InterestIntervalCustomDays,
		&updated.GracePeriodDays,
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
	updated.InterestIntervalMode = normalizeTabsInterestIntervalMode(updated.InterestIntervalMode)
	updated.InterestIntervalCustomDays = normalizeTabsInterestIntervalCustomDays(updated.InterestIntervalCustomDays)
	updated.GracePeriodDays = normalizeTabsGracePeriodDays(updated.GracePeriodDays)
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

func normalizeTabsInterestIntervalMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "daily", "bi-daily", "weekly", "bi-weekly", "monthly", "custom":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return DefaultTabsModuleSettings().InterestIntervalMode
	}
}

func normalizeTabsInterestIntervalCustomDays(value int) int {
	switch {
	case value < 1:
		return 1
	case value > 30:
		return 30
	default:
		return value
	}
}

func normalizeTabsGracePeriodDays(value int) int {
	switch {
	case value < 1:
		return 1
	case value > 30:
		return 30
	default:
		return value
	}
}

func ResolveTabsInterestEveryDays(mode string, customDays int) int {
	mode = normalizeTabsInterestIntervalMode(mode)
	customDays = normalizeTabsInterestIntervalCustomDays(customDays)

	switch mode {
	case "daily":
		return 1
	case "bi-daily":
		return 2
	case "weekly":
		return 7
	case "bi-weekly":
		return 14
	case "monthly":
		return 30
	case "custom":
		return customDays
	default:
		return 7
	}
}

func ResolveTabsGracePeriodDays(days int) int {
	return normalizeTabsGracePeriodDays(days)
}

// Backward-compatible normalizers used by user_tabs interest math.
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

func normalizeTabsInterestStartDelayValue(value int) int {
	switch {
	case value < 1:
		return 1
	case value > 3650:
		return 3650
	default:
		return value
	}
}
