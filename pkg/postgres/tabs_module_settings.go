package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type TabsModuleSettings struct {
	Enabled                 bool
	InterestRatePct         float64
	InterestEveryDays       int
	InterestStartDelayMode  string
	InterestStartDelayValue int
	InterestStartDelayUnit  string
	UpdatedBy               string
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

type TabsModuleSettingsStore struct {
	client *Client
}

func NewTabsModuleSettingsStore(client *Client) *TabsModuleSettingsStore {
	return &TabsModuleSettingsStore{client: client}
}

func DefaultTabsModuleSettings() TabsModuleSettings {
	return TabsModuleSettings{
		Enabled:                 true,
		InterestRatePct:         0,
		InterestEveryDays:       7,
		InterestStartDelayMode:  "week",
		InterestStartDelayValue: 1,
		InterestStartDelayUnit:  "weeks",
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
	interest_start_delay_mode,
	interest_start_delay_value,
	interest_start_delay_unit,
	updated_by,
	created_at,
	updated_at
)
VALUES (1, $1, $2, $3, $4, $5, $6, '', NOW(), NOW())
ON CONFLICT (id) DO NOTHING
`,
		defaults.Enabled,
		normalizeTabsInterestRate(defaults.InterestRatePct),
		normalizeTabsInterestEveryDays(defaults.InterestEveryDays),
		normalizeTabsInterestStartDelayMode(defaults.InterestStartDelayMode),
		normalizeTabsInterestStartDelayValue(defaults.InterestStartDelayValue),
		normalizeTabsInterestStartDelayUnit(defaults.InterestStartDelayUnit),
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
	interest_start_delay_mode,
	interest_start_delay_value,
	interest_start_delay_unit,
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
		&settings.InterestStartDelayMode,
		&settings.InterestStartDelayValue,
		&settings.InterestStartDelayUnit,
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
	settings.InterestStartDelayMode = normalizeTabsInterestStartDelayMode(settings.InterestStartDelayMode)
	settings.InterestStartDelayValue = normalizeTabsInterestStartDelayValue(settings.InterestStartDelayValue)
	settings.InterestStartDelayUnit = normalizeTabsInterestStartDelayUnit(settings.InterestStartDelayUnit)
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
	interest_start_delay_mode = $4,
	interest_start_delay_value = $5,
	interest_start_delay_unit = $6,
	updated_by = $7,
	updated_at = NOW()
WHERE id = 1
RETURNING
	enabled,
	interest_rate_percent,
	interest_every_days,
	interest_start_delay_mode,
	interest_start_delay_value,
	interest_start_delay_unit,
	updated_by,
	created_at,
	updated_at
`,
		settings.Enabled,
		normalizeTabsInterestRate(settings.InterestRatePct),
		normalizeTabsInterestEveryDays(settings.InterestEveryDays),
		normalizeTabsInterestStartDelayMode(settings.InterestStartDelayMode),
		normalizeTabsInterestStartDelayValue(settings.InterestStartDelayValue),
		normalizeTabsInterestStartDelayUnit(settings.InterestStartDelayUnit),
		strings.TrimSpace(settings.UpdatedBy),
	).Scan(
		&updated.Enabled,
		&updated.InterestRatePct,
		&updated.InterestEveryDays,
		&updated.InterestStartDelayMode,
		&updated.InterestStartDelayValue,
		&updated.InterestStartDelayUnit,
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
	updated.InterestStartDelayMode = normalizeTabsInterestStartDelayMode(updated.InterestStartDelayMode)
	updated.InterestStartDelayValue = normalizeTabsInterestStartDelayValue(updated.InterestStartDelayValue)
	updated.InterestStartDelayUnit = normalizeTabsInterestStartDelayUnit(updated.InterestStartDelayUnit)
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

func normalizeTabsInterestStartDelayMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "day", "week", "month", "custom":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return DefaultTabsModuleSettings().InterestStartDelayMode
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

func normalizeTabsInterestStartDelayUnit(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "days", "weeks", "months":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return DefaultTabsModuleSettings().InterestStartDelayUnit
	}
}

func ResolveTabsInterestStartDelayDays(mode string, value int, unit string) int {
	mode = normalizeTabsInterestStartDelayMode(mode)
	value = normalizeTabsInterestStartDelayValue(value)
	unit = normalizeTabsInterestStartDelayUnit(unit)

	switch mode {
	case "day":
		return 1
	case "week":
		return 7
	case "month":
		return 30
	case "custom":
		multiplier := 1
		switch unit {
		case "weeks":
			multiplier = 7
		case "months":
			multiplier = 30
		}
		delayDays := value * multiplier
		if delayDays < 1 {
			return 1
		}
		if delayDays > 3650 {
			return 3650
		}
		return delayDays
	default:
		return 7
	}
}
