package dashboard

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
)

type tabsModuleResponse struct {
	Enabled                 bool    `json:"enabled"`
	InterestRatePct         float64 `json:"interest_rate_percent"`
	InterestEveryDays       int     `json:"interest_every_days"`
	InterestStartDelayMode  string  `json:"interest_start_delay_mode"`
	InterestStartDelayValue int     `json:"interest_start_delay_value"`
	InterestStartDelayUnit  string  `json:"interest_start_delay_unit"`
}

func (h handler) tabsModule(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getTabsModule(w, r)
	case http.MethodPut:
		h.updateTabsModule(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet+", "+http.MethodPut)
	}
}

func (h handler) getTabsModule(w http.ResponseWriter, r *http.Request) {
	if _, err := h.requireEditorFeatureAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.TabsModule == nil {
		http.Error(w, "tabs module settings are not configured", http.StatusInternalServerError)
		return
	}

	if err := h.appState.TabsModule.EnsureDefault(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	settings, err := h.appState.TabsModule.Get(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if settings == nil {
		defaults := postgres.DefaultTabsModuleSettings()
		settings = &defaults
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tabsModuleToResponse(*settings))
}

func (h handler) updateTabsModule(w http.ResponseWriter, r *http.Request) {
	userSession, err := h.requireEditorFeatureAccess(r)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.TabsModule == nil {
		http.Error(w, "tabs module settings are not configured", http.StatusInternalServerError)
		return
	}

	if err := h.appState.TabsModule.EnsureDefault(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var request tabsModuleResponse
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid tabs module payload", http.StatusBadRequest)
		return
	}

	updated, err := h.appState.TabsModule.Update(r.Context(), postgres.TabsModuleSettings{
		Enabled:                 request.Enabled,
		InterestRatePct:         request.InterestRatePct,
		InterestEveryDays:       request.InterestEveryDays,
		InterestStartDelayMode:  strings.TrimSpace(request.InterestStartDelayMode),
		InterestStartDelayValue: request.InterestStartDelayValue,
		InterestStartDelayUnit:  strings.TrimSpace(request.InterestStartDelayUnit),
		UpdatedBy:               strings.TrimSpace(userSession.Login),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if updated == nil {
		http.Error(w, "tabs module settings not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tabsModuleToResponse(*updated))
}

func tabsModuleToResponse(settings postgres.TabsModuleSettings) tabsModuleResponse {
	return tabsModuleResponse{
		Enabled:                 settings.Enabled,
		InterestRatePct:         settings.InterestRatePct,
		InterestEveryDays:       settings.InterestEveryDays,
		InterestStartDelayMode:  settings.InterestStartDelayMode,
		InterestStartDelayValue: settings.InterestStartDelayValue,
		InterestStartDelayUnit:  settings.InterestStartDelayUnit,
	}
}
