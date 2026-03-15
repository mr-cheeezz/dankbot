package dashboard

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
)

type killswitchResponse struct {
	KillswitchEnabled bool `json:"killswitch_enabled"`
}

func (h handler) killswitch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}

	userSession, err := h.dashboardSession(r)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.Postgres == nil {
		http.Error(w, "bot state is not configured", http.StatusInternalServerError)
		return
	}

	stateStore := postgres.NewBotStateStore(h.appState.Postgres)
	updatedBy := strings.TrimSpace(userSession.UserID)
	if updatedBy == "" {
		updatedBy = strings.TrimSpace(userSession.Login)
	}

	state, err := stateStore.ToggleKillswitch(r.Context(), updatedBy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if h.appState.AuditLogs != nil && state != nil {
		detail := "turned killswitch off from dashboard"
		if state.KillswitchEnabled {
			detail = "turned killswitch on from dashboard"
		}

		actorName := strings.TrimSpace(userSession.DisplayName)
		if actorName == "" {
			actorName = strings.TrimSpace(userSession.Login)
		}

		_, _ = h.appState.AuditLogs.Create(r.Context(), postgres.AuditLog{
			Platform:  "web",
			ActorID:   strings.TrimSpace(userSession.UserID),
			ActorName: actorName,
			Command:   "dashboard:killswitch",
			Detail:    detail,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(killswitchResponse{
		KillswitchEnabled: state != nil && state.KillswitchEnabled,
	})
}
