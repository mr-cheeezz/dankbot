package dashboard

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
)

type unlinkIntegrationRequest struct {
	Provider string `json:"provider"`
	Target   string `json:"target"`
}

func (h handler) unlinkIntegration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}

	if _, err := h.requireIntegrationsAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil {
		http.Error(w, "dashboard state is not configured", http.StatusInternalServerError)
		return
	}

	var payload unlinkIntegrationRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid unlink payload", http.StatusBadRequest)
		return
	}

	provider := strings.TrimSpace(strings.ToLower(payload.Provider))
	target := strings.TrimSpace(strings.ToLower(payload.Target))

	var err error
	switch provider {
	case "twitch":
		switch target {
		case "streamer":
			if h.appState.TwitchAccounts == nil {
				http.Error(w, "twitch accounts are not configured", http.StatusInternalServerError)
				return
			}
			err = h.appState.TwitchAccounts.Delete(r.Context(), postgres.TwitchAccountKindStreamer)
		case "bot":
			if h.appState.TwitchAccounts == nil {
				http.Error(w, "twitch accounts are not configured", http.StatusInternalServerError)
				return
			}
			err = h.appState.TwitchAccounts.Delete(r.Context(), postgres.TwitchAccountKindBot)
		default:
			http.Error(w, "invalid twitch unlink target", http.StatusBadRequest)
			return
		}
	case "spotify":
		if h.appState.SpotifyAccounts == nil {
			http.Error(w, "spotify accounts are not configured", http.StatusInternalServerError)
			return
		}
		err = h.appState.SpotifyAccounts.Delete(r.Context(), postgres.SpotifyAccountKindStreamer)
	case "roblox":
		if h.appState.RobloxAccounts == nil {
			http.Error(w, "roblox accounts are not configured", http.StatusInternalServerError)
			return
		}
		err = h.appState.RobloxAccounts.Delete(r.Context(), postgres.RobloxAccountKindStreamer)
	case "streamlabs":
		if h.appState.StreamlabsAccounts == nil {
			http.Error(w, "streamlabs accounts are not configured", http.StatusInternalServerError)
			return
		}
		err = h.appState.StreamlabsAccounts.Delete(r.Context(), postgres.StreamlabsAccountKindStreamer)
	case "streamelements":
		if h.appState.StreamElementsAccounts == nil {
			http.Error(w, "streamelements accounts are not configured", http.StatusInternalServerError)
			return
		}
		err = h.appState.StreamElementsAccounts.Delete(r.Context(), postgres.StreamElementsAccountKindStreamer)
	case "discord":
		if h.appState.DiscordBotInstallation == nil {
			http.Error(w, "discord bot installation is not configured", http.StatusInternalServerError)
			return
		}
		err = h.appState.DiscordBotInstallation.Delete(r.Context())
	default:
		http.Error(w, "unsupported integration unlink target", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, "failed to unlink integration", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":       true,
		"provider": provider,
		"target":   target,
	})
}
