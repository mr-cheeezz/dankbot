package dashboard

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
)

type alertsResponse struct {
	Items []postgres.AlertSettingEntry `json:"items"`
}

func (h handler) alerts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getAlerts(w, r)
	case http.MethodPut:
		h.updateAlerts(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet+", "+http.MethodPut)
	}
}

func (h handler) getAlerts(w http.ResponseWriter, r *http.Request) {
	if err := h.requireDashboardAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.AlertSettings == nil {
		http.Error(w, "alerts settings are not configured", http.StatusInternalServerError)
		return
	}
	if err := h.appState.AlertSettings.EnsureDefault(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	settings, err := h.appState.AlertSettings.Get(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	entries := []postgres.AlertSettingEntry{}
	if settings != nil && settings.Entries != nil {
		entries = settings.Entries
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(alertsResponse{
		Items: entries,
	})
}

func (h handler) updateAlerts(w http.ResponseWriter, r *http.Request) {
	userSession, err := h.requireEditorFeatureAccess(r)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.AlertSettings == nil {
		http.Error(w, "alerts settings are not configured", http.StatusInternalServerError)
		return
	}

	var request alertsResponse
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid alerts payload", http.StatusBadRequest)
		return
	}

	updated, err := h.appState.AlertSettings.Update(r.Context(), request.Items, userSession.Login)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if updated == nil {
		http.Error(w, "alert settings not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(alertsResponse{
		Items: updated.Entries,
	})
}
