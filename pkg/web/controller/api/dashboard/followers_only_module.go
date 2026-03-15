package dashboard

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
)

type followersOnlyModuleResponse struct {
	Enabled                 bool `json:"enabled"`
	AutoDisableAfterMinutes int  `json:"auto_disable_after_minutes"`
}

func (h handler) followersOnlyModule(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getFollowersOnlyModule(w, r)
	case http.MethodPut:
		h.updateFollowersOnlyModule(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet+", "+http.MethodPut)
	}
}

func (h handler) getFollowersOnlyModule(w http.ResponseWriter, r *http.Request) {
	if _, err := h.requireEditorFeatureAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.FollowersOnlyModule == nil {
		http.Error(w, "followers-only module settings are not configured", http.StatusInternalServerError)
		return
	}

	if err := h.appState.FollowersOnlyModule.EnsureDefault(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	settings, err := h.appState.FollowersOnlyModule.Get(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if settings == nil {
		defaults := postgres.DefaultFollowersOnlyModuleSettings()
		settings = &defaults
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(followersOnlyModuleToResponse(*settings))
}

func (h handler) updateFollowersOnlyModule(w http.ResponseWriter, r *http.Request) {
	userSession, err := h.requireEditorFeatureAccess(r)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.FollowersOnlyModule == nil {
		http.Error(w, "followers-only module settings are not configured", http.StatusInternalServerError)
		return
	}

	if err := h.appState.FollowersOnlyModule.EnsureDefault(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var request followersOnlyModuleResponse
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid followers-only module payload", http.StatusBadRequest)
		return
	}

	updated, err := h.appState.FollowersOnlyModule.Update(r.Context(), postgres.FollowersOnlyModuleSettings{
		Enabled:                 request.Enabled,
		AutoDisableAfterMinutes: request.AutoDisableAfterMinutes,
		UpdatedBy:               strings.TrimSpace(userSession.Login),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if updated == nil {
		http.Error(w, "followers-only module settings not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(followersOnlyModuleToResponse(*updated))
}

func followersOnlyModuleToResponse(settings postgres.FollowersOnlyModuleSettings) followersOnlyModuleResponse {
	return followersOnlyModuleResponse{
		Enabled:                 settings.Enabled,
		AutoDisableAfterMinutes: settings.AutoDisableAfterMinutes,
	}
}
