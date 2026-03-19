package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type NowPlayingModuleSettings struct {
	KeywordResponse string
	UpdatedBy       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type NowPlayingModuleSettingsStore struct {
	client *Client
}

func NewNowPlayingModuleSettingsStore(client *Client) *NowPlayingModuleSettingsStore {
	return &NowPlayingModuleSettingsStore{client: client}
}

func DefaultNowPlayingModuleSettings() NowPlayingModuleSettings {
	return NowPlayingModuleSettings{
		KeywordResponse: "@{target}, use !song to see the current track. You can also use !song next or !song last.",
	}
}

func (s *NowPlayingModuleSettingsStore) EnsureDefault(ctx context.Context) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	defaults := DefaultNowPlayingModuleSettings()
	_, err = db.ExecContext(
		ctx,
		`
INSERT INTO now_playing_module_settings (
	id,
	keyword_response,
	updated_by,
	created_at,
	updated_at
)
VALUES (1, $1, '', NOW(), NOW())
ON CONFLICT (id) DO NOTHING
`,
		normalizeNowPlayingKeywordResponse(defaults.KeywordResponse),
	)
	if err != nil {
		return fmt.Errorf("ensure now playing module settings defaults: %w", err)
	}

	return nil
}

func (s *NowPlayingModuleSettingsStore) Get(ctx context.Context) (*NowPlayingModuleSettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var settings NowPlayingModuleSettings
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	keyword_response,
	updated_by,
	created_at,
	updated_at
FROM now_playing_module_settings
WHERE id = 1
`,
	).Scan(
		&settings.KeywordResponse,
		&settings.UpdatedBy,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get now playing module settings: %w", err)
	}

	settings.KeywordResponse = normalizeNowPlayingKeywordResponse(settings.KeywordResponse)
	return &settings, nil
}

func (s *NowPlayingModuleSettingsStore) Update(ctx context.Context, settings NowPlayingModuleSettings) (*NowPlayingModuleSettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var updated NowPlayingModuleSettings
	err = db.QueryRowContext(
		ctx,
		`
UPDATE now_playing_module_settings
SET
	keyword_response = $1,
	updated_by = $2,
	updated_at = NOW()
WHERE id = 1
RETURNING
	keyword_response,
	updated_by,
	created_at,
	updated_at
`,
		normalizeNowPlayingKeywordResponse(settings.KeywordResponse),
		strings.TrimSpace(settings.UpdatedBy),
	).Scan(
		&updated.KeywordResponse,
		&updated.UpdatedBy,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("update now playing module settings: %w", err)
	}

	updated.KeywordResponse = normalizeNowPlayingKeywordResponse(updated.KeywordResponse)
	return &updated, nil
}

func normalizeNowPlayingKeywordResponse(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return DefaultNowPlayingModuleSettings().KeywordResponse
	}
	return value
}
