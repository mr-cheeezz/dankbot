package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type DefaultKeywordSetting struct {
	KeywordName        string
	Enabled            bool
	AIDetectionEnabled bool
	UpdatedBy          string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type DefaultKeywordSettingStore struct {
	client *Client
}

func NewDefaultKeywordSettingStore(client *Client) *DefaultKeywordSettingStore {
	return &DefaultKeywordSettingStore{client: client}
}

func (s *DefaultKeywordSettingStore) EnsureDefaults(ctx context.Context, settings []DefaultKeywordSetting) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	for _, setting := range settings {
		setting.KeywordName = normalizeDefaultKeywordName(setting.KeywordName)
		if setting.KeywordName == "" {
			continue
		}

		if _, err := db.ExecContext(
			ctx,
			`
INSERT INTO default_keyword_settings (
	keyword_name,
	enabled,
	ai_detection_enabled,
	updated_by,
	created_at,
	updated_at
)
VALUES ($1, $2, $3, $4, NOW(), NOW())
ON CONFLICT (keyword_name) DO NOTHING
`,
			setting.KeywordName,
			setting.Enabled,
			setting.AIDetectionEnabled,
			strings.TrimSpace(setting.UpdatedBy),
		); err != nil {
			return fmt.Errorf("ensure default keyword setting %q: %w", setting.KeywordName, err)
		}
	}

	return nil
}

func (s *DefaultKeywordSettingStore) Get(ctx context.Context, keywordName string) (*DefaultKeywordSetting, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	keywordName = normalizeDefaultKeywordName(keywordName)
	if keywordName == "" {
		return nil, nil
	}

	var setting DefaultKeywordSetting
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	keyword_name,
	enabled,
	ai_detection_enabled,
	updated_by,
	created_at,
	updated_at
FROM default_keyword_settings
WHERE keyword_name = $1
`,
		keywordName,
	).Scan(
		&setting.KeywordName,
		&setting.Enabled,
		&setting.AIDetectionEnabled,
		&setting.UpdatedBy,
		&setting.CreatedAt,
		&setting.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get default keyword setting %q: %w", keywordName, err)
	}

	return &setting, nil
}

func (s *DefaultKeywordSettingStore) List(ctx context.Context) ([]DefaultKeywordSetting, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(
		ctx,
		`
SELECT
	keyword_name,
	enabled,
	ai_detection_enabled,
	updated_by,
	created_at,
	updated_at
FROM default_keyword_settings
ORDER BY keyword_name ASC
`,
	)
	if err != nil {
		return nil, fmt.Errorf("list default keyword settings: %w", err)
	}
	defer rows.Close()

	settings := make([]DefaultKeywordSetting, 0)
	for rows.Next() {
		var setting DefaultKeywordSetting
		if err := rows.Scan(
			&setting.KeywordName,
			&setting.Enabled,
			&setting.AIDetectionEnabled,
			&setting.UpdatedBy,
			&setting.CreatedAt,
			&setting.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan default keyword setting: %w", err)
		}
		settings = append(settings, setting)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate default keyword settings: %w", err)
	}

	return settings, nil
}

func (s *DefaultKeywordSettingStore) Update(ctx context.Context, setting DefaultKeywordSetting) (*DefaultKeywordSetting, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	setting.KeywordName = normalizeDefaultKeywordName(setting.KeywordName)
	if setting.KeywordName == "" {
		return nil, fmt.Errorf("keyword name is required")
	}

	var updated DefaultKeywordSetting
	err = db.QueryRowContext(
		ctx,
		`
UPDATE default_keyword_settings
SET
	enabled = $2,
	ai_detection_enabled = $3,
	updated_by = $4,
	updated_at = NOW()
WHERE keyword_name = $1
RETURNING keyword_name, enabled, ai_detection_enabled, updated_by, created_at, updated_at
`,
		setting.KeywordName,
		setting.Enabled,
		setting.AIDetectionEnabled,
		strings.TrimSpace(setting.UpdatedBy),
	).Scan(
		&updated.KeywordName,
		&updated.Enabled,
		&updated.AIDetectionEnabled,
		&updated.UpdatedBy,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("update default keyword setting %q: %w", setting.KeywordName, err)
	}

	return &updated, nil
}

func normalizeDefaultKeywordName(keywordName string) string {
	return strings.ToLower(strings.TrimSpace(keywordName))
}
