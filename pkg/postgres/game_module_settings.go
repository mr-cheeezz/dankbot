package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type GameModuleSettings struct {
	KeywordResponse         string
	PlaytimeTemplate        string
	GamesPlayedTemplate     string
	GamesPlayedItemTemplate string
	GamesPlayedLimit        int
	UpdatedBy               string
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

type GameModuleSettingsStore struct {
	client *Client
}

func NewGameModuleSettingsStore(client *Client) *GameModuleSettingsStore {
	return &GameModuleSettingsStore{client: client}
}

func DefaultGameModuleSettings() GameModuleSettings {
	return GameModuleSettings{
		KeywordResponse:         "{streamer} is currently playing {game}.",
		PlaytimeTemplate:        "{streamer} has been playing {game} for {duration}.",
		GamesPlayedTemplate:     "{label}: {items}",
		GamesPlayedItemTemplate: "{game} ({duration})",
		GamesPlayedLimit:        5,
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
	playtime_template,
	gamesplayed_template,
	gamesplayed_item_template,
	gamesplayed_limit,
	updated_by,
	created_at,
	updated_at
)
VALUES (1, $1, $2, $3, $4, $5, '', NOW(), NOW())
ON CONFLICT (id) DO NOTHING
`,
		normalizeGameKeywordResponse(defaults.KeywordResponse),
		normalizeGamePlaytimeTemplate(defaults.PlaytimeTemplate),
		normalizeGameGamesPlayedTemplate(defaults.GamesPlayedTemplate),
		normalizeGameGamesPlayedItemTemplate(defaults.GamesPlayedItemTemplate),
		normalizeGameGamesPlayedLimit(defaults.GamesPlayedLimit),
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
	playtime_template,
	gamesplayed_template,
	gamesplayed_item_template,
	gamesplayed_limit,
	updated_by,
	created_at,
	updated_at
FROM game_module_settings
WHERE id = 1
`,
	).Scan(
		&settings.KeywordResponse,
		&settings.PlaytimeTemplate,
		&settings.GamesPlayedTemplate,
		&settings.GamesPlayedItemTemplate,
		&settings.GamesPlayedLimit,
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
	settings.PlaytimeTemplate = normalizeGamePlaytimeTemplate(settings.PlaytimeTemplate)
	settings.GamesPlayedTemplate = normalizeGameGamesPlayedTemplate(settings.GamesPlayedTemplate)
	settings.GamesPlayedItemTemplate = normalizeGameGamesPlayedItemTemplate(settings.GamesPlayedItemTemplate)
	settings.GamesPlayedLimit = normalizeGameGamesPlayedLimit(settings.GamesPlayedLimit)
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
	playtime_template = $2,
	gamesplayed_template = $3,
	gamesplayed_item_template = $4,
	gamesplayed_limit = $5,
	updated_by = $6,
	updated_at = NOW()
WHERE id = 1
RETURNING
	keyword_response,
	playtime_template,
	gamesplayed_template,
	gamesplayed_item_template,
	gamesplayed_limit,
	updated_by,
	created_at,
	updated_at
`,
		normalizeGameKeywordResponse(settings.KeywordResponse),
		normalizeGamePlaytimeTemplate(settings.PlaytimeTemplate),
		normalizeGameGamesPlayedTemplate(settings.GamesPlayedTemplate),
		normalizeGameGamesPlayedItemTemplate(settings.GamesPlayedItemTemplate),
		normalizeGameGamesPlayedLimit(settings.GamesPlayedLimit),
		strings.TrimSpace(settings.UpdatedBy),
	).Scan(
		&updated.KeywordResponse,
		&updated.PlaytimeTemplate,
		&updated.GamesPlayedTemplate,
		&updated.GamesPlayedItemTemplate,
		&updated.GamesPlayedLimit,
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
	updated.PlaytimeTemplate = normalizeGamePlaytimeTemplate(updated.PlaytimeTemplate)
	updated.GamesPlayedTemplate = normalizeGameGamesPlayedTemplate(updated.GamesPlayedTemplate)
	updated.GamesPlayedItemTemplate = normalizeGameGamesPlayedItemTemplate(updated.GamesPlayedItemTemplate)
	updated.GamesPlayedLimit = normalizeGameGamesPlayedLimit(updated.GamesPlayedLimit)
	return &updated, nil
}

func normalizeGameKeywordResponse(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return DefaultGameModuleSettings().KeywordResponse
	}
	if strings.EqualFold(value, "@{target}, use !game to see what {streamer} is currently playing.") {
		return DefaultGameModuleSettings().KeywordResponse
	}
	return value
}

func normalizeGamePlaytimeTemplate(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return DefaultGameModuleSettings().PlaytimeTemplate
	}
	return value
}

func normalizeGameGamesPlayedTemplate(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return DefaultGameModuleSettings().GamesPlayedTemplate
	}
	return value
}

func normalizeGameGamesPlayedItemTemplate(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return DefaultGameModuleSettings().GamesPlayedItemTemplate
	}
	return value
}

func normalizeGameGamesPlayedLimit(value int) int {
	if value < 1 {
		return DefaultGameModuleSettings().GamesPlayedLimit
	}
	if value > 25 {
		return 25
	}
	return value
}
