package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type PromoLink struct {
	Label string `json:"label"`
	Href  string `json:"href"`
}

type PublicHomeSettings struct {
	ShowNowPlaying                  bool
	ShowNowPlayingAlbumArt          bool
	ShowNowPlayingProgress          bool
	ShowNowPlayingLinks             bool
	CommandPrefix                   string
	PromoLinks                      []PromoLink
	RobloxLinkCommandTarget         string
	RobloxLinkCommandTemplate       string
	RobloxLinkCommandDeleteTemplate string
	UpdatedBy                       string
	CreatedAt                       time.Time
	UpdatedAt                       time.Time
}

type PublicHomeSettingsStore struct {
	client *Client
}

func NewPublicHomeSettingsStore(client *Client) *PublicHomeSettingsStore {
	return &PublicHomeSettingsStore{client: client}
}

func DefaultPublicHomeSettings() PublicHomeSettings {
	return PublicHomeSettings{
		ShowNowPlaying:                  true,
		ShowNowPlayingAlbumArt:          true,
		ShowNowPlayingProgress:          true,
		ShowNowPlayingLinks:             true,
		CommandPrefix:                   "!",
		PromoLinks:                      []PromoLink{},
		RobloxLinkCommandTarget:         "dankbot",
		RobloxLinkCommandTemplate:       "",
		RobloxLinkCommandDeleteTemplate: "",
	}
}

func (s *PublicHomeSettingsStore) EnsureDefault(ctx context.Context) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	defaults := DefaultPublicHomeSettings()
	promoLinksJSON := mustMarshalPromoLinks(defaults.PromoLinks)
	_, err = db.ExecContext(
		ctx,
		`
INSERT INTO public_home_settings (
	id,
	show_now_playing,
	show_now_playing_album_art,
	show_now_playing_progress,
	show_now_playing_links,
	command_prefix,
	promo_links_json,
	roblox_link_command_target,
	roblox_link_command_template,
	roblox_link_command_delete_template,
	updated_by,
	created_at,
	updated_at
)
VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, $9, '', NOW(), NOW())
ON CONFLICT (id) DO NOTHING
`,
		defaults.ShowNowPlaying,
		defaults.ShowNowPlayingAlbumArt,
		defaults.ShowNowPlayingProgress,
		defaults.ShowNowPlayingLinks,
		normalizeCommandPrefix(defaults.CommandPrefix),
		promoLinksJSON,
		defaults.RobloxLinkCommandTarget,
		defaults.RobloxLinkCommandTemplate,
		defaults.RobloxLinkCommandDeleteTemplate,
	)
	if err != nil {
		return fmt.Errorf("ensure public home settings defaults: %w", err)
	}

	return nil
}

func (s *PublicHomeSettingsStore) Get(ctx context.Context) (*PublicHomeSettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var (
		settings       PublicHomeSettings
		promoLinksJSON string
	)
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	show_now_playing,
	show_now_playing_album_art,
	show_now_playing_progress,
	show_now_playing_links,
	command_prefix,
	promo_links_json,
	roblox_link_command_target,
	roblox_link_command_template,
	roblox_link_command_delete_template,
	updated_by,
	created_at,
	updated_at
FROM public_home_settings
WHERE id = 1
`,
	).Scan(
		&settings.ShowNowPlaying,
		&settings.ShowNowPlayingAlbumArt,
		&settings.ShowNowPlayingProgress,
		&settings.ShowNowPlayingLinks,
		&settings.CommandPrefix,
		&promoLinksJSON,
		&settings.RobloxLinkCommandTarget,
		&settings.RobloxLinkCommandTemplate,
		&settings.RobloxLinkCommandDeleteTemplate,
		&settings.UpdatedBy,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get public home settings: %w", err)
	}
	settings.CommandPrefix = normalizeCommandPrefix(settings.CommandPrefix)
	settings.PromoLinks = decodePromoLinks(promoLinksJSON)

	return &settings, nil
}

func (s *PublicHomeSettingsStore) Update(ctx context.Context, settings PublicHomeSettings) (*PublicHomeSettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var (
		updated        PublicHomeSettings
		promoLinksJSON = mustMarshalPromoLinks(settings.PromoLinks)
	)
	err = db.QueryRowContext(
		ctx,
		`
UPDATE public_home_settings
SET
	show_now_playing = $1,
	show_now_playing_album_art = $2,
	show_now_playing_progress = $3,
	show_now_playing_links = $4,
	command_prefix = $5,
	promo_links_json = $6,
	roblox_link_command_target = $7,
	roblox_link_command_template = $8,
	roblox_link_command_delete_template = $9,
	updated_by = $10,
	updated_at = NOW()
WHERE id = 1
RETURNING
	show_now_playing,
	show_now_playing_album_art,
	show_now_playing_progress,
	show_now_playing_links,
	command_prefix,
	promo_links_json,
	roblox_link_command_target,
	roblox_link_command_template,
	roblox_link_command_delete_template,
	updated_by,
	created_at,
	updated_at
`,
		settings.ShowNowPlaying,
		settings.ShowNowPlayingAlbumArt,
		settings.ShowNowPlayingProgress,
		settings.ShowNowPlayingLinks,
		normalizeCommandPrefix(settings.CommandPrefix),
		promoLinksJSON,
		normalizeLinkCommandTarget(settings.RobloxLinkCommandTarget),
		strings.TrimSpace(settings.RobloxLinkCommandTemplate),
		strings.TrimSpace(settings.RobloxLinkCommandDeleteTemplate),
		strings.TrimSpace(settings.UpdatedBy),
	).Scan(
		&updated.ShowNowPlaying,
		&updated.ShowNowPlayingAlbumArt,
		&updated.ShowNowPlayingProgress,
		&updated.ShowNowPlayingLinks,
		&updated.CommandPrefix,
		&promoLinksJSON,
		&updated.RobloxLinkCommandTarget,
		&updated.RobloxLinkCommandTemplate,
		&updated.RobloxLinkCommandDeleteTemplate,
		&updated.UpdatedBy,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("update public home settings: %w", err)
	}
	updated.CommandPrefix = normalizeCommandPrefix(updated.CommandPrefix)
	updated.PromoLinks = decodePromoLinks(promoLinksJSON)

	return &updated, nil
}

func normalizeLinkCommandTarget(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "nightbot", "fossabot", "pajbot", "custom":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return "dankbot"
	}
}

func normalizeCommandPrefix(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "!"
	}

	trimmed = strings.ReplaceAll(trimmed, " ", "")
	if trimmed == "" {
		return "!"
	}

	runes := []rune(trimmed)
	if len(runes) > 5 {
		return string(runes[:5])
	}

	return trimmed
}

func sanitizePromoLinks(items []PromoLink) []PromoLink {
	sanitized := make([]PromoLink, 0, len(items))
	for _, item := range items {
		label := strings.TrimSpace(item.Label)
		href := strings.TrimSpace(item.Href)
		if label == "" || href == "" {
			continue
		}
		if _, err := url.ParseRequestURI(href); err != nil {
			continue
		}
		sanitized = append(sanitized, PromoLink{
			Label: label,
			Href:  href,
		})
		if len(sanitized) >= 6 {
			break
		}
	}

	return sanitized
}

func mustMarshalPromoLinks(items []PromoLink) string {
	payload, err := json.Marshal(sanitizePromoLinks(items))
	if err != nil {
		return "[]"
	}

	return string(payload)
}

func decodePromoLinks(raw string) []PromoLink {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []PromoLink{}
	}

	var items []PromoLink
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return []PromoLink{}
	}

	return sanitizePromoLinks(items)
}
