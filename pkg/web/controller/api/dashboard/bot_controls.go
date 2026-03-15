package dashboard

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	modesmodule "github.com/mr-cheeezz/dankbot/pkg/modules/modes"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
)

type botModeOptionResponse struct {
	Key   string `json:"key"`
	Title string `json:"title"`
}

type botControlsResponse struct {
	CurrentModeKey string                  `json:"current_mode_key"`
	Modes          []botModeOptionResponse `json:"modes"`
}

type updateBotControlsRequest struct {
	ModeKey string `json:"mode_key"`
}

func (h handler) botControls(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getBotControls(w, r)
	case http.MethodPut:
		h.updateBotControls(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet+", "+http.MethodPut)
	}
}

func (h handler) getBotControls(w http.ResponseWriter, r *http.Request) {
	if err := h.requireDashboardAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	response, err := h.buildBotControlsResponse(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (h handler) updateBotControls(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "bot controls are not configured", http.StatusInternalServerError)
		return
	}

	var request updateBotControlsRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid bot controls payload", http.StatusBadRequest)
		return
	}

	modeStore := postgres.NewBotModeStore(h.appState.Postgres)
	stateStore := postgres.NewBotStateStore(h.appState.Postgres)

	modeKey := strings.TrimSpace(strings.ToLower(request.ModeKey))
	if modeKey == "" {
		http.Error(w, "mode_key is required", http.StatusBadRequest)
		return
	}

	mode, err := modeStore.Get(r.Context(), modeKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if mode == nil {
		http.Error(w, "mode not found", http.StatusNotFound)
		return
	}

	if err := stateStore.Ensure(r.Context(), "join"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	updatedBy := strings.TrimSpace(userSession.UserID)
	if updatedBy == "" {
		updatedBy = strings.TrimSpace(userSession.Login)
	}
	if err := stateStore.SetCurrentMode(r.Context(), mode.ModeKey, "", updatedBy); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if h.appState.AuditLogs != nil {
		actorName := strings.TrimSpace(userSession.DisplayName)
		if actorName == "" {
			actorName = strings.TrimSpace(userSession.Login)
		}

		_, _ = h.appState.AuditLogs.Create(r.Context(), postgres.AuditLog{
			Platform:  "web",
			ActorID:   strings.TrimSpace(userSession.UserID),
			ActorName: actorName,
			Command:   "dashboard:mode",
			Detail:    fmt.Sprintf("turned on %s mode from dashboard", mode.ModeKey),
		})
	}

	response, err := h.buildBotControlsResponse(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (h handler) buildBotControlsResponse(ctx context.Context) (botControlsResponse, error) {
	if h.appState == nil || h.appState.Postgres == nil {
		return botControlsResponse{}, fmt.Errorf("bot controls are not configured")
	}

	modeStore := postgres.NewBotModeStore(h.appState.Postgres)
	stateStore := postgres.NewBotStateStore(h.appState.Postgres)

	if err := modeStore.EnsureDefaults(ctx, modesmodule.BuiltInModes()); err != nil {
		return botControlsResponse{}, err
	}

	items, err := modeStore.List(ctx)
	if err != nil {
		return botControlsResponse{}, err
	}

	response := botControlsResponse{
		CurrentModeKey: "join",
		Modes:          make([]botModeOptionResponse, 0, len(items)),
	}

	if state, err := stateStore.Get(ctx); err == nil && state != nil && strings.TrimSpace(state.CurrentModeKey) != "" {
		response.CurrentModeKey = strings.TrimSpace(state.CurrentModeKey)
	}

	for _, item := range items {
		title := strings.TrimSpace(item.Title)
		if title == "" {
			title = item.ModeKey
		}
		response.Modes = append(response.Modes, botModeOptionResponse{
			Key:   item.ModeKey,
			Title: title,
		})
	}

	return response, nil
}
