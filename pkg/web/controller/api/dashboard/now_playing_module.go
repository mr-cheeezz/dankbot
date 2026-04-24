package dashboard

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
)

const builtInSongKeywordName = "what song is this"

type nowPlayingModuleResponse struct {
	Enabled                   bool   `json:"enabled"`
	AIDetectionEnabled        bool   `json:"ai_detection_enabled"`
	KeywordResponse           string `json:"keyword_response"`
	SongChangeMessageTemplate string `json:"song_change_message_template"`
	SongCommandEnabled        bool   `json:"song_command_enabled"`
	SongNextCommandEnabled    bool   `json:"song_next_command_enabled"`
	SongLastCommandEnabled    bool   `json:"song_last_command_enabled"`
}

func (h handler) nowPlayingModule(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getNowPlayingModule(w, r)
	case http.MethodPut:
		h.updateNowPlayingModule(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet+", "+http.MethodPut)
	}
}

func (h handler) getNowPlayingModule(w http.ResponseWriter, r *http.Request) {
	if _, err := h.requireEditorFeatureAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.NowPlayingModule == nil || h.appState.DefaultKeywords == nil {
		http.Error(w, "now playing module settings are not configured", http.StatusInternalServerError)
		return
	}

	if err := h.appState.NowPlayingModule.EnsureDefault(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.appState.DefaultKeywords.EnsureDefaults(r.Context(), []postgres.DefaultKeywordSetting{
		{
			KeywordName:        builtInSongKeywordName,
			Enabled:            true,
			AIDetectionEnabled: true,
		},
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	settings, err := h.appState.NowPlayingModule.Get(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if settings == nil {
		defaults := postgres.DefaultNowPlayingModuleSettings()
		settings = &defaults
	}

	keywordSetting, err := h.appState.DefaultKeywords.Get(r.Context(), builtInSongKeywordName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if keywordSetting == nil {
		defaults := postgres.DefaultKeywordSetting{
			KeywordName:        builtInSongKeywordName,
			Enabled:            true,
			AIDetectionEnabled: true,
		}
		keywordSetting = &defaults
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(nowPlayingModuleToResponse(*settings, *keywordSetting))
}

func (h handler) updateNowPlayingModule(w http.ResponseWriter, r *http.Request) {
	userSession, err := h.requireEditorFeatureAccess(r)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.NowPlayingModule == nil || h.appState.DefaultKeywords == nil {
		http.Error(w, "now playing module settings are not configured", http.StatusInternalServerError)
		return
	}

	if err := h.appState.NowPlayingModule.EnsureDefault(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.appState.DefaultKeywords.EnsureDefaults(r.Context(), []postgres.DefaultKeywordSetting{
		{
			KeywordName:        builtInSongKeywordName,
			Enabled:            true,
			AIDetectionEnabled: true,
		},
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var request nowPlayingModuleResponse
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid now playing module payload", http.StatusBadRequest)
		return
	}

	updatedSettings, err := h.appState.NowPlayingModule.Update(r.Context(), postgres.NowPlayingModuleSettings{
		KeywordResponse:           strings.TrimSpace(request.KeywordResponse),
		SongChangeMessageTemplate: strings.TrimSpace(request.SongChangeMessageTemplate),
		SongCommandEnabled:        request.SongCommandEnabled,
		SongNextCommandEnabled:    request.SongNextCommandEnabled,
		SongLastCommandEnabled:    request.SongLastCommandEnabled,
		UpdatedBy:                 strings.TrimSpace(userSession.Login),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if updatedSettings == nil {
		http.Error(w, "now playing module settings not found", http.StatusNotFound)
		return
	}

	updatedKeywordSetting, err := h.appState.DefaultKeywords.Update(r.Context(), postgres.DefaultKeywordSetting{
		KeywordName:        builtInSongKeywordName,
		Enabled:            request.Enabled,
		AIDetectionEnabled: request.AIDetectionEnabled,
		UpdatedBy:          strings.TrimSpace(userSession.Login),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(nowPlayingModuleToResponse(*updatedSettings, *updatedKeywordSetting))
}

func nowPlayingModuleToResponse(
	settings postgres.NowPlayingModuleSettings,
	keywordSetting postgres.DefaultKeywordSetting,
) nowPlayingModuleResponse {
	return nowPlayingModuleResponse{
		Enabled:                   keywordSetting.Enabled,
		AIDetectionEnabled:        keywordSetting.AIDetectionEnabled,
		KeywordResponse:           settings.KeywordResponse,
		SongChangeMessageTemplate: settings.SongChangeMessageTemplate,
		SongCommandEnabled:        settings.SongCommandEnabled,
		SongNextCommandEnabled:    settings.SongNextCommandEnabled,
		SongLastCommandEnabled:    settings.SongLastCommandEnabled,
	}
}
