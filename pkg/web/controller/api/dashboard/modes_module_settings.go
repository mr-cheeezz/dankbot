package dashboard

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
)

type modesModuleSettingsResponse struct {
	LegacyCommandsEnabled bool `json:"legacy_commands_enabled"`
}

func (h handler) modesModuleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getModesModuleSettings(w, r)
	case http.MethodPut:
		h.updateModesModuleSettings(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet+", "+http.MethodPut)
	}
}

func (h handler) getModesModuleSettings(w http.ResponseWriter, r *http.Request) {
	if _, err := h.requireEditorFeatureAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.ModesModule == nil {
		http.Error(w, "modes module settings are not configured", http.StatusInternalServerError)
		return
	}

	if err := h.appState.ModesModule.EnsureDefault(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	settings, err := h.appState.ModesModule.Get(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if settings == nil {
		defaults := postgres.DefaultModesModuleSettings()
		settings = &defaults
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(modesModuleSettingsResponse{
		LegacyCommandsEnabled: settings.LegacyCommandsEnabled,
	})
}

func (h handler) updateModesModuleSettings(w http.ResponseWriter, r *http.Request) {
	userSession, err := h.requireEditorFeatureAccess(r)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.ModesModule == nil {
		http.Error(w, "modes module settings are not configured", http.StatusInternalServerError)
		return
	}

	if err := h.appState.ModesModule.EnsureDefault(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var request modesModuleSettingsResponse
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid modes module payload", http.StatusBadRequest)
		return
	}

	updated, err := h.appState.ModesModule.Update(r.Context(), postgres.ModesModuleSettings{
		LegacyCommandsEnabled: request.LegacyCommandsEnabled,
		UpdatedBy:             strings.TrimSpace(userSession.Login),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if updated == nil {
		http.Error(w, "modes module settings not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(modesModuleSettingsResponse{
		LegacyCommandsEnabled: updated.LegacyCommandsEnabled,
	})
}
