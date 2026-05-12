package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type GameModuleSettings struct {
	Enabled                   bool
	KeywordResponse           string
	GameCommandEnabled        bool
	PlaytimeCommandEnabled    bool
	GamesPlayedCommandEnabled bool
	PlaytimeTemplate          string
	GamesPlayedTemplate       string
	GamesPlayedItemTemplate   string
	GamesPlayedLimit          int
	UpdatedBy                 string
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}

type GameModuleSettingsStore struct {
	client *Client
}

func NewGameModuleSettingsStore(client *Client) *GameModuleSettingsStore {
	return &GameModuleSettingsStore{client: client}
}

func DefaultGameModuleSettings() GameModuleSettings {
	return GameModuleSettings{
		Enabled:                   true,
		KeywordResponse:           "{streamer} is currently playing {game}.",
		GameCommandEnabled:        true,
		PlaytimeCommandEnabled:    true,
		GamesPlayedCommandEnabled: true,
		PlaytimeTemplate:          "{streamer} has been playing {game} for {duration}.",
		GamesPlayedTemplate:       "{label}: {items}",
		GamesPlayedItemTemplate:   "{game} ({duration})",
		GamesPlayedLimit:          5,
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
	enabled,
	keyword_response,
	game_command_enabled,
	playtime_command_enabled,
	gamesplayed_command_enabled,
	playtime_template,
	gamesplayed_template,
	gamesplayed_item_template,
	gamesplayed_limit,
	updated_by,
	created_at,
	updated_at
)
VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, $9, '', NOW(), NOW())
ON CONFLICT (id) DO NOTHING
`,
		defaults.Enabled,
		normalizeGameKeywordResponse(defaults.KeywordResponse),
		defaults.GameCommandEnabled,
		defaults.PlaytimeCommandEnabled,
		defaults.GamesPlayedCommandEnabled,
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
	enabled,
	keyword_response,
	game_command_enabled,
	playtime_command_enabled,
	gamesplayed_command_enabled,
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
		&settings.Enabled,
		&settings.KeywordResponse,
		&settings.GameCommandEnabled,
		&settings.PlaytimeCommandEnabled,
		&settings.GamesPlayedCommandEnabled,
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
	settings.Enabled = normalizeGameEnabled(settings.Enabled)
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
	enabled = $1,
	keyword_response = $2,
	game_command_enabled = $3,
	playtime_command_enabled = $4,
	gamesplayed_command_enabled = $5,
	playtime_template = $6,
	gamesplayed_template = $7,
	gamesplayed_item_template = $8,
	gamesplayed_limit = $9,
	updated_by = $10,
	updated_at = NOW()
WHERE id = 1
RETURNING
	enabled,
	keyword_response,
	game_command_enabled,
	playtime_command_enabled,
	gamesplayed_command_enabled,
	playtime_template,
	gamesplayed_template,
	gamesplayed_item_template,
	gamesplayed_limit,
	updated_by,
	created_at,
	updated_at
`,
		normalizeGameEnabled(settings.Enabled),
		normalizeGameKeywordResponse(settings.KeywordResponse),
		settings.GameCommandEnabled,
		settings.PlaytimeCommandEnabled,
		settings.GamesPlayedCommandEnabled,
		normalizeGamePlaytimeTemplate(settings.PlaytimeTemplate),
		normalizeGameGamesPlayedTemplate(settings.GamesPlayedTemplate),
		normalizeGameGamesPlayedItemTemplate(settings.GamesPlayedItemTemplate),
		normalizeGameGamesPlayedLimit(settings.GamesPlayedLimit),
		strings.TrimSpace(settings.UpdatedBy),
	).Scan(
		&updated.Enabled,
		&updated.KeywordResponse,
		&updated.GameCommandEnabled,
		&updated.PlaytimeCommandEnabled,
		&updated.GamesPlayedCommandEnabled,
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
	updated.Enabled = normalizeGameEnabled(updated.Enabled)
	updated.PlaytimeTemplate = normalizeGamePlaytimeTemplate(updated.PlaytimeTemplate)
	updated.GamesPlayedTemplate = normalizeGameGamesPlayedTemplate(updated.GamesPlayedTemplate)
	updated.GamesPlayedItemTemplate = normalizeGameGamesPlayedItemTemplate(updated.GamesPlayedItemTemplate)
	updated.GamesPlayedLimit = normalizeGameGamesPlayedLimit(updated.GamesPlayedLimit)
	return &updated, nil
}

func normalizeGameEnabled(value bool) bool {
	return value
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
