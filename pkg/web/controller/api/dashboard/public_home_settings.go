package dashboard

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
)

type publicHomeSettingsResponse struct {
	ShowNowPlaying            bool                `json:"show_now_playing"`
	ShowNowPlayingAlbumArt    bool                `json:"show_now_playing_album_art"`
	ShowNowPlayingProgress    bool                `json:"show_now_playing_progress"`
	ShowNowPlayingLinks       bool                `json:"show_now_playing_links"`
	CommandPrefix             string              `json:"command_prefix"`
	PromoLinks                []promoLinkResponse `json:"promo_links"`
	RobloxLinkCommandTarget   string              `json:"roblox_link_command_target"`
	RobloxLinkCommandTemplate string              `json:"roblox_link_command_template"`
}

type promoLinkResponse struct {
	Label string `json:"label"`
	Href  string `json:"href"`
}

func (h handler) publicHomeSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getPublicHomeSettings(w, r)
	case http.MethodPut:
		h.updatePublicHomeSettings(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet+", "+http.MethodPut)
	}
}

func (h handler) getPublicHomeSettings(w http.ResponseWriter, r *http.Request) {
	if err := h.requireDashboardAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.PublicHomeSettings == nil {
		http.Error(w, "public home settings are not configured", http.StatusInternalServerError)
		return
	}

	if err := h.appState.PublicHomeSettings.EnsureDefault(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	settings, err := h.appState.PublicHomeSettings.Get(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if settings == nil {
		defaults := postgres.DefaultPublicHomeSettings()
		settings = &defaults
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(publicHomeSettingsToResponse(*settings))
}

func (h handler) updatePublicHomeSettings(w http.ResponseWriter, r *http.Request) {
	userSession, err := h.dashboardSession(r)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.PublicHomeSettings == nil {
		http.Error(w, "public home settings are not configured", http.StatusInternalServerError)
		return
	}

	if err := h.appState.PublicHomeSettings.EnsureDefault(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var request publicHomeSettingsResponse
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid public home settings payload", http.StatusBadRequest)
		return
	}

	updated, err := h.appState.PublicHomeSettings.Update(r.Context(), postgres.PublicHomeSettings{
		ShowNowPlaying:            request.ShowNowPlaying,
		ShowNowPlayingAlbumArt:    request.ShowNowPlayingAlbumArt,
		ShowNowPlayingProgress:    request.ShowNowPlayingProgress,
		ShowNowPlayingLinks:       request.ShowNowPlayingLinks,
		CommandPrefix:             request.CommandPrefix,
		PromoLinks:                promoLinksFromResponse(request.PromoLinks),
		RobloxLinkCommandTarget:   request.RobloxLinkCommandTarget,
		RobloxLinkCommandTemplate: request.RobloxLinkCommandTemplate,
		UpdatedBy:                 strings.TrimSpace(userSession.Login),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if updated == nil {
		http.Error(w, "public home settings not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(publicHomeSettingsToResponse(*updated))
}

func publicHomeSettingsToResponse(settings postgres.PublicHomeSettings) publicHomeSettingsResponse {
	return publicHomeSettingsResponse{
		ShowNowPlaying:            settings.ShowNowPlaying,
		ShowNowPlayingAlbumArt:    settings.ShowNowPlayingAlbumArt,
		ShowNowPlayingProgress:    settings.ShowNowPlayingProgress,
		ShowNowPlayingLinks:       settings.ShowNowPlayingLinks,
		CommandPrefix:             settings.CommandPrefix,
		PromoLinks:                promoLinksToResponse(settings.PromoLinks),
		RobloxLinkCommandTarget:   settings.RobloxLinkCommandTarget,
		RobloxLinkCommandTemplate: settings.RobloxLinkCommandTemplate,
	}
}

func promoLinksToResponse(items []postgres.PromoLink) []promoLinkResponse {
	out := make([]promoLinkResponse, 0, len(items))
	for _, item := range items {
		out = append(out, promoLinkResponse{
			Label: item.Label,
			Href:  item.Href,
		})
	}

	return out
}

func promoLinksFromResponse(items []promoLinkResponse) []postgres.PromoLink {
	out := make([]postgres.PromoLink, 0, len(items))
	for _, item := range items {
		out = append(out, postgres.PromoLink{
			Label: item.Label,
			Href:  item.Href,
		})
	}

	return out
}
