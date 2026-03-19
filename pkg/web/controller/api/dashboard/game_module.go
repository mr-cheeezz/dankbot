package dashboard

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
)

const builtInGameKeywordName = "what game is this"

type gameModuleResponse struct {
	Enabled                 bool   `json:"enabled"`
	AIDetectionEnabled      bool   `json:"ai_detection_enabled"`
	KeywordResponse         string `json:"keyword_response"`
	PlaytimeTemplate        string `json:"playtime_template"`
	GamesPlayedTemplate     string `json:"gamesplayed_template"`
	GamesPlayedItemTemplate string `json:"gamesplayed_item_template"`
	GamesPlayedLimit        int    `json:"gamesplayed_limit"`
}

func (h handler) gameModule(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getGameModule(w, r)
	case http.MethodPut:
		h.updateGameModule(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet+", "+http.MethodPut)
	}
}

func (h handler) getGameModule(w http.ResponseWriter, r *http.Request) {
	if _, err := h.requireEditorFeatureAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.GameModule == nil || h.appState.DefaultKeywords == nil {
		http.Error(w, "game module settings are not configured", http.StatusInternalServerError)
		return
	}

	if err := h.appState.GameModule.EnsureDefault(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.appState.DefaultKeywords.EnsureDefaults(r.Context(), []postgres.DefaultKeywordSetting{
		{
			KeywordName:        builtInGameKeywordName,
			Enabled:            true,
			AIDetectionEnabled: true,
		},
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	settings, err := h.appState.GameModule.Get(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if settings == nil {
		defaults := postgres.DefaultGameModuleSettings()
		settings = &defaults
	}

	keywordSetting, err := h.appState.DefaultKeywords.Get(r.Context(), builtInGameKeywordName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if keywordSetting == nil {
		defaults := postgres.DefaultKeywordSetting{
			KeywordName:        builtInGameKeywordName,
			Enabled:            true,
			AIDetectionEnabled: true,
		}
		keywordSetting = &defaults
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(gameModuleToResponse(*settings, *keywordSetting))
}

func (h handler) updateGameModule(w http.ResponseWriter, r *http.Request) {
	userSession, err := h.requireEditorFeatureAccess(r)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.GameModule == nil || h.appState.DefaultKeywords == nil {
		http.Error(w, "game module settings are not configured", http.StatusInternalServerError)
		return
	}

	if err := h.appState.GameModule.EnsureDefault(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.appState.DefaultKeywords.EnsureDefaults(r.Context(), []postgres.DefaultKeywordSetting{
		{
			KeywordName:        builtInGameKeywordName,
			Enabled:            true,
			AIDetectionEnabled: true,
		},
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var request gameModuleResponse
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid game module payload", http.StatusBadRequest)
		return
	}

	updatedSettings, err := h.appState.GameModule.Update(r.Context(), postgres.GameModuleSettings{
		KeywordResponse:         strings.TrimSpace(request.KeywordResponse),
		PlaytimeTemplate:        strings.TrimSpace(request.PlaytimeTemplate),
		GamesPlayedTemplate:     strings.TrimSpace(request.GamesPlayedTemplate),
		GamesPlayedItemTemplate: strings.TrimSpace(request.GamesPlayedItemTemplate),
		GamesPlayedLimit:        request.GamesPlayedLimit,
		UpdatedBy:               strings.TrimSpace(userSession.Login),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if updatedSettings == nil {
		http.Error(w, "game module settings not found", http.StatusNotFound)
		return
	}

	updatedKeywordSetting, err := h.appState.DefaultKeywords.Update(r.Context(), postgres.DefaultKeywordSetting{
		KeywordName:        builtInGameKeywordName,
		Enabled:            request.Enabled,
		AIDetectionEnabled: request.AIDetectionEnabled,
		UpdatedBy:          strings.TrimSpace(userSession.Login),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(gameModuleToResponse(*updatedSettings, *updatedKeywordSetting))
}

func gameModuleToResponse(
	settings postgres.GameModuleSettings,
	keywordSetting postgres.DefaultKeywordSetting,
) gameModuleResponse {
	return gameModuleResponse{
		Enabled:                 keywordSetting.Enabled,
		AIDetectionEnabled:      keywordSetting.AIDetectionEnabled,
		KeywordResponse:         settings.KeywordResponse,
		PlaytimeTemplate:        settings.PlaytimeTemplate,
		GamesPlayedTemplate:     settings.GamesPlayedTemplate,
		GamesPlayedItemTemplate: settings.GamesPlayedItemTemplate,
		GamesPlayedLimit:        settings.GamesPlayedLimit,
	}
}
