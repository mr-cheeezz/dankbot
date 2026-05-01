package dashboard

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
)

type rustLogModuleResponse struct {
	Enabled bool `json:"enabled"`
}

func (h handler) rustLogModule(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getRustLogModule(w, r)
	case http.MethodPut:
		h.updateRustLogModule(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet+", "+http.MethodPut)
	}
}

func (h handler) getRustLogModule(w http.ResponseWriter, r *http.Request) {
	if _, err := h.requireEditorFeatureAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.RustLogModule == nil || h.appState.Config == nil {
		http.Error(w, "rustlog module settings are not configured", http.StatusInternalServerError)
		return
	}

	if err := h.appState.RustLogModule.EnsureDefault(r.Context(), h.appState.Config.RustLog.Enabled); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	settings, err := h.appState.RustLogModule.Get(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if settings == nil {
		defaults := postgres.DefaultRustLogModuleSettings(h.appState.Config.RustLog.Enabled)
		settings = &defaults
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(rustLogModuleResponse{
		Enabled: settings.Enabled,
	})
}

func (h handler) updateRustLogModule(w http.ResponseWriter, r *http.Request) {
	userSession, err := h.requireEditorFeatureAccess(r)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.RustLogModule == nil || h.appState.Config == nil {
		http.Error(w, "rustlog module settings are not configured", http.StatusInternalServerError)
		return
	}

	if err := h.appState.RustLogModule.EnsureDefault(r.Context(), h.appState.Config.RustLog.Enabled); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var request rustLogModuleResponse
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid rustlog module payload", http.StatusBadRequest)
		return
	}

	updated, err := h.appState.RustLogModule.Update(r.Context(), postgres.RustLogModuleSettings{
		Enabled:   request.Enabled,
		UpdatedBy: strings.TrimSpace(userSession.Login),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if updated == nil {
		http.Error(w, "rustlog module settings not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(rustLogModuleResponse{
		Enabled: updated.Enabled,
	})
}
