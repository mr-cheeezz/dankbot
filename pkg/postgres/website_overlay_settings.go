package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type WebsiteOverlaySettings struct {
	PollsEnabled         bool
	PredictionsEnabled   bool
	PollPosition         string
	PredictionPosition   string
	PollOffsetX          int
	PollOffsetY          int
	PredictionOffsetX    int
	PredictionOffsetY    int
	PollScale            float64
	PollBarColor         string
	PollTitleColor       string
	PollTextColor        string
	PredictionScale      float64
	PredictionLeftColor  string
	PredictionRightColor string
	PredictionTextColor  string
	PredictionTrackColor string
	UpdatedBy            string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type WebsiteOverlaySettingsStore struct {
	client *Client
}

func NewWebsiteOverlaySettingsStore(client *Client) *WebsiteOverlaySettingsStore {
	return &WebsiteOverlaySettingsStore{client: client}
}

func DefaultWebsiteOverlaySettings() WebsiteOverlaySettings {
	return WebsiteOverlaySettings{
		PollsEnabled:         true,
		PredictionsEnabled:   true,
		PollPosition:         "top-right",
		PredictionPosition:   "top-center",
		PollOffsetX:          24,
		PollOffsetY:          24,
		PredictionOffsetX:    24,
		PredictionOffsetY:    24,
		PollScale:            1,
		PollBarColor:         "#7dd3fc",
		PollTitleColor:       "#ffffff",
		PollTextColor:        "#f8fafc",
		PredictionScale:      1,
		PredictionLeftColor:  "#60a5fa",
		PredictionRightColor: "#f472b6",
		PredictionTextColor:  "#ffffff",
		PredictionTrackColor: "rgba(15, 23, 42, 0.28)",
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
	poll_scale,
	poll_bar_color,
	poll_title_color,
	poll_text_color,
	prediction_scale,
	prediction_left_color,
	prediction_right_color,
	prediction_text_color,
	prediction_track_color,
	updated_by,
	created_at,
	updated_at
)
VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, '', NOW(), NOW())
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
		defaults.PollScale,
		defaults.PollBarColor,
		defaults.PollTitleColor,
		defaults.PollTextColor,
		defaults.PredictionScale,
		defaults.PredictionLeftColor,
		defaults.PredictionRightColor,
		defaults.PredictionTextColor,
		defaults.PredictionTrackColor,
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
	poll_scale,
	poll_bar_color,
	poll_title_color,
	poll_text_color,
	prediction_scale,
	prediction_left_color,
	prediction_right_color,
	prediction_text_color,
	prediction_track_color,
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
		&settings.PollScale,
		&settings.PollBarColor,
		&settings.PollTitleColor,
		&settings.PollTextColor,
		&settings.PredictionScale,
		&settings.PredictionLeftColor,
		&settings.PredictionRightColor,
		&settings.PredictionTextColor,
		&settings.PredictionTrackColor,
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
	settings.PollScale = normalizeOverlayScale(settings.PollScale)
	settings.PollBarColor = normalizeOverlayColor(settings.PollBarColor, "#7dd3fc")
	settings.PollTitleColor = normalizeOverlayColor(settings.PollTitleColor, "#ffffff")
	settings.PollTextColor = normalizeOverlayColor(settings.PollTextColor, "#f8fafc")
	settings.PredictionScale = normalizeOverlayScale(settings.PredictionScale)
	settings.PredictionLeftColor = normalizeOverlayColor(settings.PredictionLeftColor, "#60a5fa")
	settings.PredictionRightColor = normalizeOverlayColor(settings.PredictionRightColor, "#f472b6")
	settings.PredictionTextColor = normalizeOverlayColor(settings.PredictionTextColor, "#ffffff")
	settings.PredictionTrackColor = normalizeOverlayColor(settings.PredictionTrackColor, "rgba(15, 23, 42, 0.28)")

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
	poll_scale = $9,
	poll_bar_color = $10,
	poll_title_color = $11,
	poll_text_color = $12,
	prediction_scale = $13,
	prediction_left_color = $14,
	prediction_right_color = $15,
	prediction_text_color = $16,
	prediction_track_color = $17,
	updated_by = $18,
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
	poll_scale,
	poll_bar_color,
	poll_title_color,
	poll_text_color,
	prediction_scale,
	prediction_left_color,
	prediction_right_color,
	prediction_text_color,
	prediction_track_color,
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
		normalizeOverlayScale(settings.PollScale),
		normalizeOverlayColor(settings.PollBarColor, "#7dd3fc"),
		normalizeOverlayColor(settings.PollTitleColor, "#ffffff"),
		normalizeOverlayColor(settings.PollTextColor, "#f8fafc"),
		normalizeOverlayScale(settings.PredictionScale),
		normalizeOverlayColor(settings.PredictionLeftColor, "#60a5fa"),
		normalizeOverlayColor(settings.PredictionRightColor, "#f472b6"),
		normalizeOverlayColor(settings.PredictionTextColor, "#ffffff"),
		normalizeOverlayColor(settings.PredictionTrackColor, "rgba(15, 23, 42, 0.28)"),
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
		&updated.PollScale,
		&updated.PollBarColor,
		&updated.PollTitleColor,
		&updated.PollTextColor,
		&updated.PredictionScale,
		&updated.PredictionLeftColor,
		&updated.PredictionRightColor,
		&updated.PredictionTextColor,
		&updated.PredictionTrackColor,
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
	case "top-left", "top-right", "top-center", "bottom-left", "bottom-right":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return "top-right"
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

func normalizeOverlayScale(raw float64) float64 {
	if raw < 0.5 {
		return 0.5
	}
	if raw > 2 {
		return 2
	}
	return raw
}

var overlayHexColorPattern = regexp.MustCompile(`^#(?:[0-9a-fA-F]{6}|[0-9a-fA-F]{3})$`)

func normalizeOverlayColor(raw, fallback string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return fallback
	}
	if overlayHexColorPattern.MatchString(value) {
		return strings.ToLower(value)
	}
	if strings.HasPrefix(strings.ToLower(value), "rgba(") && strings.HasSuffix(value, ")") {
		return normalizeRGBAColor(value, fallback)
	}
	return fallback
}

func normalizeRGBAColor(raw, fallback string) string {
	trimmed := strings.TrimSpace(raw)
	parts := strings.Split(strings.TrimSuffix(strings.TrimPrefix(trimmed, "rgba("), ")"), ",")
	if len(parts) != 4 {
		return fallback
	}
	values := make([]string, 0, 4)
	for index, part := range parts {
		value := strings.TrimSpace(part)
		if index < 3 {
			number, err := strconv.Atoi(value)
			if err != nil || number < 0 || number > 255 {
				return fallback
			}
			values = append(values, strconv.Itoa(number))
			continue
		}
		alphaValue, err := strconv.ParseFloat(value, 64)
		if err != nil || alphaValue < 0 || alphaValue > 1 {
			return fallback
		}
		values = append(values, strconv.FormatFloat(alphaValue, 'f', -1, 64))
	}
	return "rgba(" + strings.Join(values, ", ") + ")"
}
