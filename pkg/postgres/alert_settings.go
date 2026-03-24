package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type AlertSettingEntry struct {
	ID            string `json:"id"`
	Provider      string `json:"provider"`
	Section       string `json:"section"`
	Label         string `json:"label"`
	Source        string `json:"source"`
	Behavior      string `json:"behavior"`
	Status        string `json:"status"`
	Enabled       bool   `json:"enabled"`
	Template      string `json:"template"`
	Scope         string `json:"scope"`
	Note          string `json:"note,omitempty"`
	MinimumLabel  string `json:"minimum_label,omitempty"`
	MinimumValue  *int   `json:"minimum_value,omitempty"`
	MinimumUnit   string `json:"minimum_unit,omitempty"`
	MinimumPrefix string `json:"minimum_prefix,omitempty"`
}

type AlertSettings struct {
	Entries   []AlertSettingEntry
	UpdatedBy string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AlertSettingsStore struct {
	client *Client
}

func NewAlertSettingsStore(client *Client) *AlertSettingsStore {
	return &AlertSettingsStore{client: client}
}

func (s *AlertSettingsStore) EnsureDefault(ctx context.Context) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(
		ctx,
		`
INSERT INTO alert_settings (
	id,
	entries_json,
	updated_by,
	created_at,
	updated_at
)
VALUES (1, '[]'::jsonb, '', NOW(), NOW())
ON CONFLICT (id) DO NOTHING
`,
	)
	if err != nil {
		return fmt.Errorf("ensure alert settings defaults: %w", err)
	}

	return nil
}

func (s *AlertSettingsStore) Get(ctx context.Context) (*AlertSettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var (
		entriesJSON []byte
		settings    AlertSettings
	)
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	entries_json,
	updated_by,
	created_at,
	updated_at
FROM alert_settings
WHERE id = 1
`,
	).Scan(
		&entriesJSON,
		&settings.UpdatedBy,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get alert settings: %w", err)
	}

	if len(entriesJSON) > 0 {
		if err := json.Unmarshal(entriesJSON, &settings.Entries); err != nil {
			return nil, fmt.Errorf("decode alert settings entries: %w", err)
		}
	}
	if settings.Entries == nil {
		settings.Entries = []AlertSettingEntry{}
	}

	return &settings, nil
}

func (s *AlertSettingsStore) Update(ctx context.Context, entries []AlertSettingEntry, updatedBy string) (*AlertSettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	if entries == nil {
		entries = []AlertSettingEntry{}
	}
	normalized := make([]AlertSettingEntry, 0, len(entries))
	for _, entry := range entries {
		id := strings.TrimSpace(entry.ID)
		if id == "" {
			continue
		}
		normalized = append(normalized, AlertSettingEntry{
			ID:            id,
			Provider:      strings.TrimSpace(entry.Provider),
			Section:       strings.TrimSpace(entry.Section),
			Label:         strings.TrimSpace(entry.Label),
			Source:        strings.TrimSpace(entry.Source),
			Behavior:      strings.TrimSpace(entry.Behavior),
			Status:        strings.TrimSpace(entry.Status),
			Enabled:       entry.Enabled,
			Template:      strings.TrimSpace(entry.Template),
			Scope:         strings.TrimSpace(entry.Scope),
			Note:          strings.TrimSpace(entry.Note),
			MinimumLabel:  strings.TrimSpace(entry.MinimumLabel),
			MinimumValue:  entry.MinimumValue,
			MinimumUnit:   strings.TrimSpace(entry.MinimumUnit),
			MinimumPrefix: strings.TrimSpace(entry.MinimumPrefix),
		})
	}

	payload, err := json.Marshal(normalized)
	if err != nil {
		return nil, fmt.Errorf("encode alert settings entries: %w", err)
	}

	var (
		updatedJSON []byte
		updated     AlertSettings
	)
	err = db.QueryRowContext(
		ctx,
		`
UPDATE alert_settings
SET
	entries_json = $1::jsonb,
	updated_by = $2,
	updated_at = NOW()
WHERE id = 1
RETURNING
	entries_json,
	updated_by,
	created_at,
	updated_at
`,
		string(payload),
		strings.TrimSpace(updatedBy),
	).Scan(
		&updatedJSON,
		&updated.UpdatedBy,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("update alert settings: %w", err)
	}

	if len(updatedJSON) > 0 {
		if err := json.Unmarshal(updatedJSON, &updated.Entries); err != nil {
			return nil, fmt.Errorf("decode updated alert settings entries: %w", err)
		}
	}
	if updated.Entries == nil {
		updated.Entries = []AlertSettingEntry{}
	}

	return &updated, nil
}
