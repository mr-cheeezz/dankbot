package dashboard

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
)

type websiteOverlaySettingsResponse struct {
	PollsEnabled         bool    `json:"polls_enabled"`
	PredictionsEnabled   bool    `json:"predictions_enabled"`
	PollPosition         string  `json:"poll_position"`
	PredictionPosition   string  `json:"prediction_position"`
	PollOffsetX          int     `json:"poll_offset_x"`
	PollOffsetY          int     `json:"poll_offset_y"`
	PredictionOffsetX    int     `json:"prediction_offset_x"`
	PredictionOffsetY    int     `json:"prediction_offset_y"`
	PollScale            float64 `json:"poll_scale"`
	PollBarColor         string  `json:"poll_bar_color"`
	PollTitleColor       string  `json:"poll_title_color"`
	PollTextColor        string  `json:"poll_text_color"`
	PredictionScale      float64 `json:"prediction_scale"`
	PredictionLeftColor  string  `json:"prediction_left_color"`
	PredictionRightColor string  `json:"prediction_right_color"`
	PredictionTextColor  string  `json:"prediction_text_color"`
	PredictionTrackColor string  `json:"prediction_track_color"`
}

func (h handler) websiteOverlaySettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getWebsiteOverlaySettings(w, r)
	case http.MethodPut:
		h.updateWebsiteOverlaySettings(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet+", "+http.MethodPut)
	}
}

func (h handler) getWebsiteOverlaySettings(w http.ResponseWriter, r *http.Request) {
	if err := h.requireDashboardAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.WebsiteOverlaySettings == nil {
		http.Error(w, "website overlay settings are not configured", http.StatusInternalServerError)
		return
	}

	if err := h.appState.WebsiteOverlaySettings.EnsureDefault(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	settings, err := h.appState.WebsiteOverlaySettings.Get(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if settings == nil {
		defaults := postgres.DefaultWebsiteOverlaySettings()
		settings = &defaults
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(websiteOverlaySettingsToResponse(*settings))
}

func (h handler) updateWebsiteOverlaySettings(w http.ResponseWriter, r *http.Request) {
	userSession, err := h.requireEditorFeatureAccess(r)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.WebsiteOverlaySettings == nil {
		http.Error(w, "website overlay settings are not configured", http.StatusInternalServerError)
		return
	}

	if err := h.appState.WebsiteOverlaySettings.EnsureDefault(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var request websiteOverlaySettingsResponse
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid website overlay settings payload", http.StatusBadRequest)
		return
	}

	updated, err := h.appState.WebsiteOverlaySettings.Update(r.Context(), postgres.WebsiteOverlaySettings{
		PollsEnabled:         request.PollsEnabled,
		PredictionsEnabled:   request.PredictionsEnabled,
		PollPosition:         request.PollPosition,
		PredictionPosition:   request.PredictionPosition,
		PollOffsetX:          request.PollOffsetX,
		PollOffsetY:          request.PollOffsetY,
		PredictionOffsetX:    request.PredictionOffsetX,
		PredictionOffsetY:    request.PredictionOffsetY,
		PollScale:            request.PollScale,
		PollBarColor:         request.PollBarColor,
		PollTitleColor:       request.PollTitleColor,
		PollTextColor:        request.PollTextColor,
		PredictionScale:      request.PredictionScale,
		PredictionLeftColor:  request.PredictionLeftColor,
		PredictionRightColor: request.PredictionRightColor,
		PredictionTextColor:  request.PredictionTextColor,
		PredictionTrackColor: request.PredictionTrackColor,
		UpdatedBy:            strings.TrimSpace(userSession.Login),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if updated == nil {
		http.Error(w, "website overlay settings not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(websiteOverlaySettingsToResponse(*updated))
}

func websiteOverlaySettingsToResponse(settings postgres.WebsiteOverlaySettings) websiteOverlaySettingsResponse {
	return websiteOverlaySettingsResponse{
		PollsEnabled:         settings.PollsEnabled,
		PredictionsEnabled:   settings.PredictionsEnabled,
		PollPosition:         settings.PollPosition,
		PredictionPosition:   settings.PredictionPosition,
		PollOffsetX:          settings.PollOffsetX,
		PollOffsetY:          settings.PollOffsetY,
		PredictionOffsetX:    settings.PredictionOffsetX,
		PredictionOffsetY:    settings.PredictionOffsetY,
		PollScale:            settings.PollScale,
		PollBarColor:         settings.PollBarColor,
		PollTitleColor:       settings.PollTitleColor,
		PollTextColor:        settings.PollTextColor,
		PredictionScale:      settings.PredictionScale,
		PredictionLeftColor:  settings.PredictionLeftColor,
		PredictionRightColor: settings.PredictionRightColor,
		PredictionTextColor:  settings.PredictionTextColor,
		PredictionTrackColor: settings.PredictionTrackColor,
	}
}
