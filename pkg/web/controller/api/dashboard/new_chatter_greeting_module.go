package dashboard

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
)

type newChatterGreetingModuleResponse struct {
	Enabled  bool     `json:"enabled"`
	Messages []string `json:"messages"`
}

func (h handler) newChatterGreetingModule(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getNewChatterGreetingModule(w, r)
	case http.MethodPut:
		h.updateNewChatterGreetingModule(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet+", "+http.MethodPut)
	}
}

func (h handler) getNewChatterGreetingModule(w http.ResponseWriter, r *http.Request) {
	if _, err := h.requireEditorFeatureAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.NewChatterGreeting == nil {
		http.Error(w, "new chatter greeting module settings are not configured", http.StatusInternalServerError)
		return
	}

	if err := h.appState.NewChatterGreeting.EnsureDefault(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	settings, err := h.appState.NewChatterGreeting.Get(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if settings == nil {
		defaults := postgres.DefaultNewChatterGreetingModuleSettings()
		settings = &defaults
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(newChatterGreetingModuleToResponse(*settings))
}

func (h handler) updateNewChatterGreetingModule(w http.ResponseWriter, r *http.Request) {
	userSession, err := h.requireEditorFeatureAccess(r)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.NewChatterGreeting == nil {
		http.Error(w, "new chatter greeting module settings are not configured", http.StatusInternalServerError)
		return
	}

	if err := h.appState.NewChatterGreeting.EnsureDefault(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var request newChatterGreetingModuleResponse
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid new chatter greeting module payload", http.StatusBadRequest)
		return
	}

	updated, err := h.appState.NewChatterGreeting.Update(r.Context(), postgres.NewChatterGreetingModuleSettings{
		Enabled:   request.Enabled,
		Messages:  append([]string(nil), request.Messages...),
		UpdatedBy: strings.TrimSpace(userSession.Login),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if updated == nil {
		http.Error(w, "new chatter greeting module settings not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(newChatterGreetingModuleToResponse(*updated))
}

func newChatterGreetingModuleToResponse(settings postgres.NewChatterGreetingModuleSettings) newChatterGreetingModuleResponse {
	return newChatterGreetingModuleResponse{
		Enabled:  settings.Enabled,
		Messages: append([]string(nil), settings.Messages...),
	}
}
