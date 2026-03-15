package dashboard

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	keywordsmodule "github.com/mr-cheeezz/dankbot/pkg/modules/keywords"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
)

type defaultKeywordResponse struct {
	KeywordName        string `json:"keyword_name"`
	Enabled            bool   `json:"enabled"`
	AIDetectionEnabled bool   `json:"ai_detection_enabled"`
}

type updateDefaultKeywordRequest struct {
	KeywordName        string `json:"keyword_name"`
	Enabled            bool   `json:"enabled"`
	AIDetectionEnabled bool   `json:"ai_detection_enabled"`
}

func (h handler) defaultKeywords(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listDefaultKeywords(w, r)
	case http.MethodPut:
		h.updateDefaultKeyword(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet+", "+http.MethodPut)
	}
}

func (h handler) listDefaultKeywords(w http.ResponseWriter, r *http.Request) {
	if err := h.requireDashboardAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.DefaultKeywords == nil {
		http.Error(w, "default keywords are not configured", http.StatusInternalServerError)
		return
	}

	if err := h.appState.DefaultKeywords.EnsureDefaults(r.Context(), keywordsmodule.DefaultSettingsDefaults()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	items, err := h.appState.DefaultKeywords.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]defaultKeywordResponse, 0, len(items))
	for _, item := range items {
		response = append(response, defaultKeywordToResponse(item))
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"items": response})
}

func (h handler) updateDefaultKeyword(w http.ResponseWriter, r *http.Request) {
	userSession, err := h.dashboardSession(r)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.DefaultKeywords == nil {
		http.Error(w, "default keywords are not configured", http.StatusInternalServerError)
		return
	}

	var request updateDefaultKeywordRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid default keyword payload", http.StatusBadRequest)
		return
	}

	keywordName := strings.ToLower(strings.TrimSpace(request.KeywordName))
	if keywordName == "" {
		http.Error(w, "keyword name is required", http.StatusBadRequest)
		return
	}

	if err := h.appState.DefaultKeywords.EnsureDefaults(r.Context(), keywordsmodule.DefaultSettingsDefaults()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	updated, err := h.appState.DefaultKeywords.Update(r.Context(), postgres.DefaultKeywordSetting{
		KeywordName:        keywordName,
		Enabled:            request.Enabled,
		AIDetectionEnabled: request.AIDetectionEnabled,
		UpdatedBy:          strings.TrimSpace(userSession.Login),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if updated == nil {
		http.Error(w, "default keyword not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(defaultKeywordToResponse(*updated))
}

func defaultKeywordToResponse(item postgres.DefaultKeywordSetting) defaultKeywordResponse {
	return defaultKeywordResponse{
		KeywordName:        item.KeywordName,
		Enabled:            item.Enabled,
		AIDetectionEnabled: item.AIDetectionEnabled,
	}
}
