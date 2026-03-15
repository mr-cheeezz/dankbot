package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type GameModuleSettings struct {
	KeywordResponse string
	UpdatedBy       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type GameModuleSettingsStore struct {
	client *Client
}

func NewGameModuleSettingsStore(client *Client) *GameModuleSettingsStore {
	return &GameModuleSettingsStore{client: client}
}

func DefaultGameModuleSettings() GameModuleSettings {
	return GameModuleSettings{
		KeywordResponse: "@{target}, use !game to see what {streamer} is currently playing.",
	}
}

func (s *GameModuleSettingsStore) EnsureDefault(ctx context.Context) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	defaults := DefaultGameModuleSettings()
	_, err = db.ExecContext(
		ctx,
		`
INSERT INTO game_module_settings (
	id,
	keyword_response,
	updated_by,
	created_at,
	updated_at
)
VALUES (1, $1, '', NOW(), NOW())
ON CONFLICT (id) DO NOTHING
`,
		normalizeGameKeywordResponse(defaults.KeywordResponse),
	)
	if err != nil {
		return fmt.Errorf("ensure game module settings defaults: %w", err)
	}

	return nil
}

func (s *GameModuleSettingsStore) Get(ctx context.Context) (*GameModuleSettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var settings GameModuleSettings
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	keyword_response,
	updated_by,
	created_at,
	updated_at
FROM game_module_settings
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
		return nil, fmt.Errorf("get game module settings: %w", err)
	}

	settings.KeywordResponse = normalizeGameKeywordResponse(settings.KeywordResponse)
	return &settings, nil
}

func (s *GameModuleSettingsStore) Update(ctx context.Context, settings GameModuleSettings) (*GameModuleSettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var updated GameModuleSettings
	err = db.QueryRowContext(
		ctx,
		`
UPDATE game_module_settings
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
		normalizeGameKeywordResponse(settings.KeywordResponse),
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
		return nil, fmt.Errorf("update game module settings: %w", err)
	}

	updated.KeywordResponse = normalizeGameKeywordResponse(updated.KeywordResponse)
	return &updated, nil
}

func normalizeGameKeywordResponse(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return DefaultGameModuleSettings().KeywordResponse
	}
	return value
}
