package dashboard

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
)

type followersOnlyModuleResponse struct {
	Enabled                 bool   `json:"enabled"`
	EnabledWhenOffline      bool   `json:"enabled_when_offline"`
	AutoDisableEnabled      bool   `json:"auto_disable_enabled"`
	AutoDisableAfterMinutes int    `json:"auto_disable_after_minutes"`
	OnlineSlowModeAction    string `json:"online_slow_mode_action"`
	OnlineSlowModeSeconds   int    `json:"online_slow_mode_seconds"`
	OnlineEmoteModeAction   string `json:"online_emote_mode_action"`
	OnlineUniqueChatAction  string `json:"online_unique_chat_mode_action"`
	OnlineSubscriberAction  string `json:"online_subscriber_mode_action"`
	OnlineFollowerAction    string `json:"online_follower_mode_action"`
	OnlineFollowerMinutes   int    `json:"online_follower_mode_minutes"`
	OfflineSlowModeAction   string `json:"offline_slow_mode_action"`
	OfflineSlowModeSeconds  int    `json:"offline_slow_mode_seconds"`
	OfflineEmoteModeAction  string `json:"offline_emote_mode_action"`
	OfflineUniqueChatAction string `json:"offline_unique_chat_mode_action"`
	OfflineSubscriberAction string `json:"offline_subscriber_mode_action"`
	OfflineFollowerAction   string `json:"offline_follower_mode_action"`
	OfflineFollowerMinutes  int    `json:"offline_follower_mode_minutes"`
}

func (h handler) followersOnlyModule(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getFollowersOnlyModule(w, r)
	case http.MethodPut:
		h.updateFollowersOnlyModule(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet+", "+http.MethodPut)
	}
}

func (h handler) getFollowersOnlyModule(w http.ResponseWriter, r *http.Request) {
	if _, err := h.requireEditorFeatureAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.FollowersOnlyModule == nil {
		http.Error(w, "followers-only module settings are not configured", http.StatusInternalServerError)
		return
	}

	if err := h.appState.FollowersOnlyModule.EnsureDefault(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	settings, err := h.appState.FollowersOnlyModule.Get(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if settings == nil {
		defaults := postgres.DefaultFollowersOnlyModuleSettings()
		settings = &defaults
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(followersOnlyModuleToResponse(*settings))
}

func (h handler) updateFollowersOnlyModule(w http.ResponseWriter, r *http.Request) {
	userSession, err := h.requireEditorFeatureAccess(r)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.FollowersOnlyModule == nil {
		http.Error(w, "followers-only module settings are not configured", http.StatusInternalServerError)
		return
	}

	if err := h.appState.FollowersOnlyModule.EnsureDefault(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var request followersOnlyModuleResponse
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid followers-only module payload", http.StatusBadRequest)
		return
	}

	updated, err := h.appState.FollowersOnlyModule.Update(r.Context(), postgres.FollowersOnlyModuleSettings{
		Enabled:                 request.Enabled,
		EnabledWhenOffline:      request.EnabledWhenOffline,
		AutoDisableEnabled:      request.AutoDisableEnabled,
		AutoDisableAfterMinutes: request.AutoDisableAfterMinutes,
		OnlineSlowModeAction:    request.OnlineSlowModeAction,
		OnlineSlowModeSeconds:   request.OnlineSlowModeSeconds,
		OnlineEmoteModeAction:   request.OnlineEmoteModeAction,
		OnlineUniqueChatAction:  request.OnlineUniqueChatAction,
		OnlineSubscriberAction:  request.OnlineSubscriberAction,
		OnlineFollowerAction:    request.OnlineFollowerAction,
		OnlineFollowerMinutes:   request.OnlineFollowerMinutes,
		OfflineSlowModeAction:   request.OfflineSlowModeAction,
		OfflineSlowModeSeconds:  request.OfflineSlowModeSeconds,
		OfflineEmoteModeAction:  request.OfflineEmoteModeAction,
		OfflineUniqueChatAction: request.OfflineUniqueChatAction,
		OfflineSubscriberAction: request.OfflineSubscriberAction,
		OfflineFollowerAction:   request.OfflineFollowerAction,
		OfflineFollowerMinutes:  request.OfflineFollowerMinutes,
		UpdatedBy:               strings.TrimSpace(userSession.Login),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if updated == nil {
		http.Error(w, "followers-only module settings not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(followersOnlyModuleToResponse(*updated))
}

func followersOnlyModuleToResponse(settings postgres.FollowersOnlyModuleSettings) followersOnlyModuleResponse {
	return followersOnlyModuleResponse{
		Enabled:                 settings.Enabled,
		EnabledWhenOffline:      settings.EnabledWhenOffline,
		AutoDisableEnabled:      settings.AutoDisableEnabled,
		AutoDisableAfterMinutes: settings.AutoDisableAfterMinutes,
		OnlineSlowModeAction:    settings.OnlineSlowModeAction,
		OnlineSlowModeSeconds:   settings.OnlineSlowModeSeconds,
		OnlineEmoteModeAction:   settings.OnlineEmoteModeAction,
		OnlineUniqueChatAction:  settings.OnlineUniqueChatAction,
		OnlineSubscriberAction:  settings.OnlineSubscriberAction,
		OnlineFollowerAction:    settings.OnlineFollowerAction,
		OnlineFollowerMinutes:   settings.OnlineFollowerMinutes,
		OfflineSlowModeAction:   settings.OfflineSlowModeAction,
		OfflineSlowModeSeconds:  settings.OfflineSlowModeSeconds,
		OfflineEmoteModeAction:  settings.OfflineEmoteModeAction,
		OfflineUniqueChatAction: settings.OfflineUniqueChatAction,
		OfflineSubscriberAction: settings.OfflineSubscriberAction,
		OfflineFollowerAction:   settings.OfflineFollowerAction,
		OfflineFollowerMinutes:  settings.OfflineFollowerMinutes,
	}
}
