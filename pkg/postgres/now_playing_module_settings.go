package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type NowPlayingModuleSettings struct {
	KeywordResponse           string
	SongChangeMessageTemplate string
	SongCommandEnabled        bool
	SongNextCommandEnabled    bool
	SongLastCommandEnabled    bool
	UpdatedBy                 string
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
}

type NowPlayingModuleSettingsStore struct {
	client *Client
}

func NewNowPlayingModuleSettingsStore(client *Client) *NowPlayingModuleSettingsStore {
	return &NowPlayingModuleSettingsStore{client: client}
}

func DefaultNowPlayingModuleSettings() NowPlayingModuleSettings {
	return NowPlayingModuleSettings{
		KeywordResponse:           "@{target}, use !song to see the current track. You can also use !song next or !song last.",
		SongChangeMessageTemplate: "{streamer} is now listening to {song} PogU",
		SongCommandEnabled:        true,
		SongNextCommandEnabled:    true,
		SongLastCommandEnabled:    true,
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
	song_change_message_template,
	song_command_enabled,
	song_next_command_enabled,
	song_last_command_enabled,
	updated_by,
	created_at,
	updated_at
)
VALUES (1, $1, $2, $3, $4, $5, '', NOW(), NOW())
ON CONFLICT (id) DO NOTHING
`,
		normalizeNowPlayingKeywordResponse(defaults.KeywordResponse),
		normalizeNowPlayingSongChangeMessageTemplate(defaults.SongChangeMessageTemplate),
		defaults.SongCommandEnabled,
		defaults.SongNextCommandEnabled,
		defaults.SongLastCommandEnabled,
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
	song_change_message_template,
	song_command_enabled,
	song_next_command_enabled,
	song_last_command_enabled,
	updated_by,
	created_at,
	updated_at
FROM now_playing_module_settings
WHERE id = 1
`,
	).Scan(
		&settings.KeywordResponse,
		&settings.SongChangeMessageTemplate,
		&settings.SongCommandEnabled,
		&settings.SongNextCommandEnabled,
		&settings.SongLastCommandEnabled,
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
	settings.SongChangeMessageTemplate = normalizeNowPlayingSongChangeMessageTemplate(settings.SongChangeMessageTemplate)
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
	song_change_message_template = $2,
	song_command_enabled = $3,
	song_next_command_enabled = $4,
	song_last_command_enabled = $5,
	updated_by = $6,
	updated_at = NOW()
WHERE id = 1
RETURNING
	keyword_response,
	song_change_message_template,
	song_command_enabled,
	song_next_command_enabled,
	song_last_command_enabled,
	updated_by,
	created_at,
	updated_at
`,
		normalizeNowPlayingKeywordResponse(settings.KeywordResponse),
		normalizeNowPlayingSongChangeMessageTemplate(settings.SongChangeMessageTemplate),
		settings.SongCommandEnabled,
		settings.SongNextCommandEnabled,
		settings.SongLastCommandEnabled,
		strings.TrimSpace(settings.UpdatedBy),
	).Scan(
		&updated.KeywordResponse,
		&updated.SongChangeMessageTemplate,
		&updated.SongCommandEnabled,
		&updated.SongNextCommandEnabled,
		&updated.SongLastCommandEnabled,
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
	updated.SongChangeMessageTemplate = normalizeNowPlayingSongChangeMessageTemplate(updated.SongChangeMessageTemplate)
	return &updated, nil
}

func normalizeNowPlayingKeywordResponse(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return DefaultNowPlayingModuleSettings().KeywordResponse
	}
	return value
}

func normalizeNowPlayingSongChangeMessageTemplate(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return DefaultNowPlayingModuleSettings().SongChangeMessageTemplate
	}
	return value
}
