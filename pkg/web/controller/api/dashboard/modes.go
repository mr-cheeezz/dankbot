package dashboard

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	modesmodule "github.com/mr-cheeezz/dankbot/pkg/modules/modes"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
)

type modeResponse struct {
	ID                     string `json:"id"`
	Key                    string `json:"key"`
	Title                  string `json:"title"`
	Description            string `json:"description"`
	KeywordName            string `json:"keyword_name"`
	KeywordDescription     string `json:"keyword_description"`
	KeywordResponse        string `json:"keyword_response"`
	CoordinatedTwitchTitle string `json:"coordinated_twitch_title"`
	TimerEnabled           bool   `json:"timer_enabled"`
	TimerMessage           string `json:"timer_message"`
	TimerIntervalSeconds   int    `json:"timer_interval_seconds"`
	Builtin                bool   `json:"builtin"`
}

type modesListResponse struct {
	Items []modeResponse `json:"items"`
}

type saveModeRequest struct {
	Key                    string `json:"key"`
	Title                  string `json:"title"`
	Description            string `json:"description"`
	KeywordName            string `json:"keyword_name"`
	KeywordDescription     string `json:"keyword_description"`
	KeywordResponse        string `json:"keyword_response"`
	CoordinatedTwitchTitle string `json:"coordinated_twitch_title"`
	TimerEnabled           bool   `json:"timer_enabled"`
	TimerMessage           string `json:"timer_message"`
	TimerIntervalSeconds   int    `json:"timer_interval_seconds"`
	OriginalKey            string `json:"original_key"`
}

func (h handler) modes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listModes(w, r)
	case http.MethodPost:
		h.createMode(w, r)
	case http.MethodPut:
		h.updateMode(w, r)
	case http.MethodDelete:
		h.deleteMode(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet+", "+http.MethodPost+", "+http.MethodPut+", "+http.MethodDelete)
	}
}

func (h handler) listModes(w http.ResponseWriter, r *http.Request) {
	if _, err := h.requireEditorFeatureAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	items, err := h.fetchModes(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(modesListResponse{Items: items})
}

func (h handler) createMode(w http.ResponseWriter, r *http.Request) {
	userSession, err := h.requireEditorFeatureAccess(r)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	modeStore, request, err := h.decodeModeSaveRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	modeKey := normalizeModeKey(request.Key)
	if modeKey == "" {
		http.Error(w, "mode key is required", http.StatusBadRequest)
		return
	}

	current, err := modeStore.Get(r.Context(), modeKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if current != nil {
		http.Error(w, "mode already exists", http.StatusConflict)
		return
	}

	mode := modeFromRequest(request)
	if err := modeStore.Save(r.Context(), mode); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.logDashboardModeChange(r, userSession, fmt.Sprintf("created %s mode from dashboard", mode.ModeKey))
	h.respondWithModes(w, r)
}

func (h handler) updateMode(w http.ResponseWriter, r *http.Request) {
	userSession, err := h.requireEditorFeatureAccess(r)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	modeStore, request, err := h.decodeModeSaveRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	originalKey := normalizeModeKey(request.OriginalKey)
	if originalKey == "" {
		originalKey = normalizeModeKey(request.Key)
	}
	if originalKey == "" {
		http.Error(w, "mode key is required", http.StatusBadRequest)
		return
	}

	current, err := modeStore.Get(r.Context(), originalKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if current == nil {
		http.Error(w, "mode not found", http.StatusNotFound)
		return
	}

	nextKey := normalizeModeKey(request.Key)
	if current.IsBuiltin {
		nextKey = current.ModeKey
	}
	if nextKey == "" {
		http.Error(w, "mode key is required", http.StatusBadRequest)
		return
	}
	if nextKey != current.ModeKey {
		http.Error(w, "mode keys cannot be renamed after creation", http.StatusBadRequest)
		return
	}

	next := modeFromRequest(request)
	next.ModeKey = current.ModeKey
	next.IsBuiltin = current.IsBuiltin
	next.LastTimerSentAt = current.LastTimerSentAt
	next.CreatedAt = current.CreatedAt
	if err := modeStore.Save(r.Context(), next); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.logDashboardModeChange(r, userSession, fmt.Sprintf("updated %s mode from dashboard", next.ModeKey))
	h.respondWithModes(w, r)
}

func (h handler) deleteMode(w http.ResponseWriter, r *http.Request) {
	userSession, err := h.requireEditorFeatureAccess(r)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.Postgres == nil {
		http.Error(w, "modes are not configured", http.StatusInternalServerError)
		return
	}

	modeKey := normalizeModeKey(r.URL.Query().Get("mode_key"))
	if modeKey == "" {
		http.Error(w, "mode_key is required", http.StatusBadRequest)
		return
	}

	modeStore := postgres.NewBotModeStore(h.appState.Postgres)
	if err := modeStore.EnsureDefaults(r.Context(), modesmodule.BuiltInModes()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	current, err := modeStore.Get(r.Context(), modeKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if current == nil {
		http.Error(w, "mode not found", http.StatusNotFound)
		return
	}
	if current.IsBuiltin {
		http.Error(w, "built-in modes cannot be deleted", http.StatusBadRequest)
		return
	}

	if err := modeStore.Delete(r.Context(), modeKey); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stateStore := postgres.NewBotStateStore(h.appState.Postgres)
	if state, err := stateStore.Get(r.Context()); err == nil && state != nil && strings.EqualFold(strings.TrimSpace(state.CurrentModeKey), modeKey) {
		updatedBy := strings.TrimSpace(userSession.UserID)
		if updatedBy == "" {
			updatedBy = strings.TrimSpace(userSession.Login)
		}
		if err := stateStore.SetCurrentMode(r.Context(), "join", "", updatedBy); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	h.logDashboardModeChange(r, userSession, fmt.Sprintf("deleted %s mode from dashboard", modeKey))
	h.respondWithModes(w, r)
}

func (h handler) fetchModes(r *http.Request) ([]modeResponse, error) {
	if h.appState == nil || h.appState.Postgres == nil {
		return nil, fmt.Errorf("modes are not configured")
	}

	modeStore := postgres.NewBotModeStore(h.appState.Postgres)
	if err := modeStore.EnsureDefaults(r.Context(), modesmodule.BuiltInModes()); err != nil {
		return nil, err
	}

	items, err := modeStore.List(r.Context())
	if err != nil {
		return nil, err
	}

	response := make([]modeResponse, 0, len(items))
	for _, item := range items {
		response = append(response, modeToResponse(item))
	}

	return response, nil
}

func (h handler) respondWithModes(w http.ResponseWriter, r *http.Request) {
	items, err := h.fetchModes(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(modesListResponse{Items: items})
}

func (h handler) decodeModeSaveRequest(r *http.Request) (*postgres.BotModeStore, saveModeRequest, error) {
	if h.appState == nil || h.appState.Postgres == nil {
		return nil, saveModeRequest{}, fmt.Errorf("modes are not configured")
	}

	var request saveModeRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, saveModeRequest{}, fmt.Errorf("invalid mode payload")
	}

	modeStore := postgres.NewBotModeStore(h.appState.Postgres)
	if err := modeStore.EnsureDefaults(r.Context(), modesmodule.BuiltInModes()); err != nil {
		return nil, saveModeRequest{}, err
	}

	return modeStore, request, nil
}

func modeToResponse(item postgres.BotMode) modeResponse {
	return modeResponse{
		ID:                     item.ModeKey,
		Key:                    item.ModeKey,
		Title:                  item.Title,
		Description:            item.Description,
		KeywordName:            item.KeywordName,
		KeywordDescription:     item.KeywordDescription,
		KeywordResponse:        item.KeywordResponse,
		CoordinatedTwitchTitle: item.CoordinatedTwitchTitle,
		TimerEnabled:           item.TimerEnabled,
		TimerMessage:           item.TimerMessage,
		TimerIntervalSeconds:   item.TimerIntervalSeconds,
		Builtin:                item.IsBuiltin,
	}
}

func modeFromRequest(request saveModeRequest) postgres.BotMode {
	timerIntervalSeconds := request.TimerIntervalSeconds
	if timerIntervalSeconds <= 0 {
		timerIntervalSeconds = 180
	}

	return postgres.BotMode{
		ModeKey:                normalizeModeKey(request.Key),
		Title:                  strings.TrimSpace(request.Title),
		Description:            strings.TrimSpace(request.Description),
		KeywordName:            strings.TrimSpace(request.KeywordName),
		KeywordDescription:     strings.TrimSpace(request.KeywordDescription),
		KeywordResponse:        strings.TrimSpace(request.KeywordResponse),
		CoordinatedTwitchTitle: strings.TrimSpace(request.CoordinatedTwitchTitle),
		IsBuiltin:              false,
		TimerEnabled:           request.TimerEnabled,
		TimerMessage:           strings.TrimSpace(request.TimerMessage),
		TimerIntervalSeconds:   timerIntervalSeconds,
	}
}

func normalizeModeKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func (h handler) logDashboardModeChange(r *http.Request, userSession *session.UserSession, detail string) {
	if h.appState == nil || h.appState.AuditLogs == nil || strings.TrimSpace(detail) == "" || userSession == nil {
		return
	}

	actorName := strings.TrimSpace(userSession.DisplayName)
	if actorName == "" {
		actorName = strings.TrimSpace(userSession.Login)
	}

	_, _ = h.appState.AuditLogs.Create(r.Context(), postgres.AuditLog{
		Platform:  "web",
		ActorID:   strings.TrimSpace(userSession.UserID),
		ActorName: actorName,
		Command:   "dashboard:modes",
		Detail:    detail,
	})
}
