package dashboard

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
)

type userProfileModuleResponse struct {
	Enabled              bool `json:"enabled"`
	ShowTabSection       bool `json:"show_tab_section"`
	ShowTabHistory       bool `json:"show_tab_history"`
	ShowRedemption       bool `json:"show_redemption_activity"`
	ShowPollStats        bool `json:"show_poll_stats"`
	ShowPredictionStats  bool `json:"show_prediction_stats"`
	ShowLastSeen         bool `json:"show_last_seen"`
	ShowLastChatActivity bool `json:"show_last_chat_activity"`
}

func (h handler) userProfileModule(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getUserProfileModule(w, r)
	case http.MethodPut:
		h.updateUserProfileModule(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet+", "+http.MethodPut)
	}
}

func (h handler) getUserProfileModule(w http.ResponseWriter, r *http.Request) {
	if _, err := h.requireEditorFeatureAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.UserProfileModule == nil {
		http.Error(w, "user profile module settings are not configured", http.StatusInternalServerError)
		return
	}

	if err := h.appState.UserProfileModule.EnsureDefault(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	settings, err := h.appState.UserProfileModule.Get(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if settings == nil {
		defaults := postgres.DefaultUserProfileModuleSettings()
		settings = &defaults
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(userProfileModuleToResponse(*settings))
}

func (h handler) updateUserProfileModule(w http.ResponseWriter, r *http.Request) {
	userSession, err := h.requireEditorFeatureAccess(r)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.UserProfileModule == nil {
		http.Error(w, "user profile module settings are not configured", http.StatusInternalServerError)
		return
	}

	if err := h.appState.UserProfileModule.EnsureDefault(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var request userProfileModuleResponse
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid user profile module payload", http.StatusBadRequest)
		return
	}

	updated, err := h.appState.UserProfileModule.Update(r.Context(), postgres.UserProfileModuleSettings{
		Enabled:              request.Enabled,
		ShowTabSection:       request.ShowTabSection,
		ShowTabHistory:       request.ShowTabHistory,
		ShowRedemption:       request.ShowRedemption,
		ShowPollStats:        request.ShowPollStats,
		ShowPredictionStats:  request.ShowPredictionStats,
		ShowLastSeen:         request.ShowLastSeen,
		ShowLastChatActivity: request.ShowLastChatActivity,
		UpdatedBy:            strings.TrimSpace(userSession.Login),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if updated == nil {
		http.Error(w, "user profile module settings not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(userProfileModuleToResponse(*updated))
}

func userProfileModuleToResponse(settings postgres.UserProfileModuleSettings) userProfileModuleResponse {
	return userProfileModuleResponse{
		Enabled:              settings.Enabled,
		ShowTabSection:       settings.ShowTabSection,
		ShowTabHistory:       settings.ShowTabHistory,
		ShowRedemption:       settings.ShowRedemption,
		ShowPollStats:        settings.ShowPollStats,
		ShowPredictionStats:  settings.ShowPredictionStats,
		ShowLastSeen:         settings.ShowLastSeen,
		ShowLastChatActivity: settings.ShowLastChatActivity,
	}
}
