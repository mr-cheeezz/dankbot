package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type WebsiteOverlaySettings struct {
	PollsEnabled       bool
	PredictionsEnabled bool
	PollPosition       string
	PredictionPosition string
	PollOffsetX        int
	PollOffsetY        int
	PredictionOffsetX  int
	PredictionOffsetY  int
	UpdatedBy          string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type WebsiteOverlaySettingsStore struct {
	client *Client
}

func NewWebsiteOverlaySettingsStore(client *Client) *WebsiteOverlaySettingsStore {
	return &WebsiteOverlaySettingsStore{client: client}
}

func DefaultWebsiteOverlaySettings() WebsiteOverlaySettings {
	return WebsiteOverlaySettings{
		PollsEnabled:       true,
		PredictionsEnabled: true,
		PollPosition:       "bottom-left",
		PredictionPosition: "bottom-right",
		PollOffsetX:        24,
		PollOffsetY:        24,
		PredictionOffsetX:  24,
		PredictionOffsetY:  24,
	}
}

func (s *WebsiteOverlaySettingsStore) EnsureDefault(ctx context.Context) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	defaults := DefaultWebsiteOverlaySettings()
	_, err = db.ExecContext(
		ctx,
		`
INSERT INTO website_overlay_settings (
	id,
	polls_enabled,
	predictions_enabled,
	poll_position,
	prediction_position,
	poll_offset_x,
	poll_offset_y,
	prediction_offset_x,
	prediction_offset_y,
	updated_by,
	created_at,
	updated_at
)
VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, '', NOW(), NOW())
ON CONFLICT (id) DO NOTHING
`,
		defaults.PollsEnabled,
		defaults.PredictionsEnabled,
		defaults.PollPosition,
		defaults.PredictionPosition,
		defaults.PollOffsetX,
		defaults.PollOffsetY,
		defaults.PredictionOffsetX,
		defaults.PredictionOffsetY,
	)
	if err != nil {
		return fmt.Errorf("ensure website overlay settings defaults: %w", err)
	}

	return nil
}

func (s *WebsiteOverlaySettingsStore) Get(ctx context.Context) (*WebsiteOverlaySettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var settings WebsiteOverlaySettings
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	polls_enabled,
	predictions_enabled,
	poll_position,
	prediction_position,
	poll_offset_x,
	poll_offset_y,
	prediction_offset_x,
	prediction_offset_y,
	updated_by,
	created_at,
	updated_at
FROM website_overlay_settings
WHERE id = 1
`,
	).Scan(
		&settings.PollsEnabled,
		&settings.PredictionsEnabled,
		&settings.PollPosition,
		&settings.PredictionPosition,
		&settings.PollOffsetX,
		&settings.PollOffsetY,
		&settings.PredictionOffsetX,
		&settings.PredictionOffsetY,
		&settings.UpdatedBy,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get website overlay settings: %w", err)
	}

	settings.PollPosition = normalizeOverlayPosition(settings.PollPosition)
	settings.PredictionPosition = normalizeOverlayPosition(settings.PredictionPosition)
	settings.PollOffsetX = normalizeOverlayOffset(settings.PollOffsetX)
	settings.PollOffsetY = normalizeOverlayOffset(settings.PollOffsetY)
	settings.PredictionOffsetX = normalizeOverlayOffset(settings.PredictionOffsetX)
	settings.PredictionOffsetY = normalizeOverlayOffset(settings.PredictionOffsetY)

	return &settings, nil
}

func (s *WebsiteOverlaySettingsStore) Update(ctx context.Context, settings WebsiteOverlaySettings) (*WebsiteOverlaySettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var updated WebsiteOverlaySettings
	err = db.QueryRowContext(
		ctx,
		`
UPDATE website_overlay_settings
SET
	polls_enabled = $1,
	predictions_enabled = $2,
	poll_position = $3,
	prediction_position = $4,
	poll_offset_x = $5,
	poll_offset_y = $6,
	prediction_offset_x = $7,
	prediction_offset_y = $8,
	updated_by = $9,
	updated_at = NOW()
WHERE id = 1
RETURNING
	polls_enabled,
	predictions_enabled,
	poll_position,
	prediction_position,
	poll_offset_x,
	poll_offset_y,
	prediction_offset_x,
	prediction_offset_y,
	updated_by,
	created_at,
	updated_at
`,
		settings.PollsEnabled,
		settings.PredictionsEnabled,
		normalizeOverlayPosition(settings.PollPosition),
		normalizeOverlayPosition(settings.PredictionPosition),
		normalizeOverlayOffset(settings.PollOffsetX),
		normalizeOverlayOffset(settings.PollOffsetY),
		normalizeOverlayOffset(settings.PredictionOffsetX),
		normalizeOverlayOffset(settings.PredictionOffsetY),
		strings.TrimSpace(settings.UpdatedBy),
	).Scan(
		&updated.PollsEnabled,
		&updated.PredictionsEnabled,
		&updated.PollPosition,
		&updated.PredictionPosition,
		&updated.PollOffsetX,
		&updated.PollOffsetY,
		&updated.PredictionOffsetX,
		&updated.PredictionOffsetY,
		&updated.UpdatedBy,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("update website overlay settings: %w", err)
	}

	return &updated, nil
}

func normalizeOverlayPosition(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "top-left", "top-right", "bottom-left", "bottom-right":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return "bottom-left"
	}
}

func normalizeOverlayOffset(raw int) int {
	if raw < 0 {
		return 0
	}
	if raw > 400 {
		return 400
	}
	return raw
}
