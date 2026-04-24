package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"
)

type SpamFilterHypeSettings struct {
	Enabled                bool
	DisableDurationSeconds int
	BitsEnabled            bool
	BitsThreshold          int
	GiftedSubsEnabled      bool
	GiftedSubsThreshold    int
	RaidsEnabled           bool
	RaidsThreshold         int
	DonationsEnabled       bool
	DonationsThreshold     float64
	DisabledFilterKeys     []string
	UpdatedBy              string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type SpamFilterHypeSettingsStore struct {
	client *Client
}

func NewSpamFilterHypeSettingsStore(client *Client) *SpamFilterHypeSettingsStore {
	return &SpamFilterHypeSettingsStore{client: client}
}

func DefaultSpamFilterHypeSettings() SpamFilterHypeSettings {
	return SpamFilterHypeSettings{
		Enabled:                false,
		DisableDurationSeconds: 180,
		BitsEnabled:            true,
		BitsThreshold:          1000,
		GiftedSubsEnabled:      true,
		GiftedSubsThreshold:    10,
		RaidsEnabled:           true,
		RaidsThreshold:         50,
		DonationsEnabled:       false,
		DonationsThreshold:     25,
		DisabledFilterKeys:     []string{},
	}
}

func (s *SpamFilterHypeSettingsStore) EnsureDefault(ctx context.Context) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	defaults := normalizeSpamFilterHypeSettings(DefaultSpamFilterHypeSettings())
	disabledRaw, err := json.Marshal(defaults.DisabledFilterKeys)
	if err != nil {
		return fmt.Errorf("marshal spam filter hype disabled keys: %w", err)
	}

	_, err = db.ExecContext(
		ctx,
		`
INSERT INTO spam_filter_hype_settings (
	id,
	enabled,
	disable_duration_seconds,
	bits_enabled,
	bits_threshold,
	gifted_subs_enabled,
	gifted_subs_threshold,
	raids_enabled,
	raids_threshold,
	donations_enabled,
	donations_threshold,
	disabled_filter_keys,
	updated_by,
	created_at,
	updated_at
)
VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, '', NOW(), NOW())
ON CONFLICT (id) DO NOTHING
`,
		defaults.Enabled,
		defaults.DisableDurationSeconds,
		defaults.BitsEnabled,
		defaults.BitsThreshold,
		defaults.GiftedSubsEnabled,
		defaults.GiftedSubsThreshold,
		defaults.RaidsEnabled,
		defaults.RaidsThreshold,
		defaults.DonationsEnabled,
		defaults.DonationsThreshold,
		disabledRaw,
	)
	if err != nil {
		return fmt.Errorf("ensure spam filter hype settings defaults: %w", err)
	}

	return nil
}

func (s *SpamFilterHypeSettingsStore) Get(ctx context.Context) (*SpamFilterHypeSettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var settings SpamFilterHypeSettings
	var disabledRaw []byte
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	enabled,
	disable_duration_seconds,
	bits_enabled,
	bits_threshold,
	gifted_subs_enabled,
	gifted_subs_threshold,
	raids_enabled,
	raids_threshold,
	donations_enabled,
	donations_threshold,
	disabled_filter_keys,
	updated_by,
	created_at,
	updated_at
FROM spam_filter_hype_settings
WHERE id = 1
`,
	).Scan(
		&settings.Enabled,
		&settings.DisableDurationSeconds,
		&settings.BitsEnabled,
		&settings.BitsThreshold,
		&settings.GiftedSubsEnabled,
		&settings.GiftedSubsThreshold,
		&settings.RaidsEnabled,
		&settings.RaidsThreshold,
		&settings.DonationsEnabled,
		&settings.DonationsThreshold,
		&disabledRaw,
		&settings.UpdatedBy,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get spam filter hype settings: %w", err)
	}

	settings.DisabledFilterKeys = normalizeSpamFilterKeyList(decodeSpamRoleList(disabledRaw))
	normalized := normalizeSpamFilterHypeSettings(settings)
	return &normalized, nil
}

func (s *SpamFilterHypeSettingsStore) Update(ctx context.Context, settings SpamFilterHypeSettings) (*SpamFilterHypeSettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	settings = normalizeSpamFilterHypeSettings(settings)
	disabledRaw, err := json.Marshal(settings.DisabledFilterKeys)
	if err != nil {
		return nil, fmt.Errorf("marshal spam filter hype disabled keys: %w", err)
	}

	var updated SpamFilterHypeSettings
	var updatedDisabledRaw []byte
	err = db.QueryRowContext(
		ctx,
		`
UPDATE spam_filter_hype_settings
SET
	enabled = $1,
	disable_duration_seconds = $2,
	bits_enabled = $3,
	bits_threshold = $4,
	gifted_subs_enabled = $5,
	gifted_subs_threshold = $6,
	raids_enabled = $7,
	raids_threshold = $8,
	donations_enabled = $9,
	donations_threshold = $10,
	disabled_filter_keys = $11,
	updated_by = $12,
	updated_at = NOW()
WHERE id = 1
RETURNING
	enabled,
	disable_duration_seconds,
	bits_enabled,
	bits_threshold,
	gifted_subs_enabled,
	gifted_subs_threshold,
	raids_enabled,
	raids_threshold,
	donations_enabled,
	donations_threshold,
	disabled_filter_keys,
	updated_by,
	created_at,
	updated_at
`,
		settings.Enabled,
		settings.DisableDurationSeconds,
		settings.BitsEnabled,
		settings.BitsThreshold,
		settings.GiftedSubsEnabled,
		settings.GiftedSubsThreshold,
		settings.RaidsEnabled,
		settings.RaidsThreshold,
		settings.DonationsEnabled,
		settings.DonationsThreshold,
		disabledRaw,
		strings.TrimSpace(settings.UpdatedBy),
	).Scan(
		&updated.Enabled,
		&updated.DisableDurationSeconds,
		&updated.BitsEnabled,
		&updated.BitsThreshold,
		&updated.GiftedSubsEnabled,
		&updated.GiftedSubsThreshold,
		&updated.RaidsEnabled,
		&updated.RaidsThreshold,
		&updated.DonationsEnabled,
		&updated.DonationsThreshold,
		&updatedDisabledRaw,
		&updated.UpdatedBy,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("update spam filter hype settings: %w", err)
	}

	updated.DisabledFilterKeys = normalizeSpamFilterKeyList(decodeSpamRoleList(updatedDisabledRaw))
	normalized := normalizeSpamFilterHypeSettings(updated)
	return &normalized, nil
}

func normalizeSpamFilterHypeSettings(settings SpamFilterHypeSettings) SpamFilterHypeSettings {
	settings.DisabledFilterKeys = normalizeSpamFilterKeyList(settings.DisabledFilterKeys)
	if settings.DisableDurationSeconds < 5 {
		settings.DisableDurationSeconds = 5
	}
	if settings.BitsThreshold < 1 {
		settings.BitsThreshold = 1
	}
	if settings.GiftedSubsThreshold < 1 {
		settings.GiftedSubsThreshold = 1
	}
	if settings.RaidsThreshold < 1 {
		settings.RaidsThreshold = 1
	}
	if math.IsNaN(settings.DonationsThreshold) || math.IsInf(settings.DonationsThreshold, 0) || settings.DonationsThreshold < 0.01 {
		settings.DonationsThreshold = 0.01
	}
	if settings.DonationsThreshold > 1000000 {
		settings.DonationsThreshold = 1000000
	}
	return settings
}

func normalizeSpamFilterKeyList(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		key := normalizeSpamFilterKey(value)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	return out
}
