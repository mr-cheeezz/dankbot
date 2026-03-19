package twitchauth

import (
	"encoding/json"
	"errors"
	"net/http"

	webaccess "github.com/mr-cheeezz/dankbot/pkg/web/access"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
	"github.com/mr-cheeezz/dankbot/pkg/web/state"
)

type handler struct {
	appState *state.State
}

type errorResponse struct {
	Error string `json:"error"`
}

func Register(mux *http.ServeMux, appState *state.State) {
	h := handler{appState: appState}

	mux.Handle("/auth/login", http.HandlerFunc(h.siteLogin))
	mux.Handle("/auth/streamer", http.HandlerFunc(h.streamerConnect))
	mux.Handle("/auth/bot", http.HandlerFunc(h.botConnect))
}

func (h handler) siteLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	url, err := h.appState.TwitchOAuth.SiteLoginURL(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}

func (h handler) streamerConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}
	if !h.requireIntegrationAccess(w, r) {
		return
	}

	url, err := h.appState.TwitchOAuth.StreamerConnectURL(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}

func (h handler) botConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}
	if !h.requireBotIntegrationAccess(w, r) {
		return
	}

	url, err := h.appState.TwitchOAuth.BotConnectURL(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: message})
}

func writeMethodNotAllowed(w http.ResponseWriter, allowedMethod string) {
	w.Header().Set("Allow", allowedMethod)
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func (h handler) requireIntegrationAccess(w http.ResponseWriter, r *http.Request) bool {
	_, access, err := webaccess.LoadDashboardSession(r.Context(), r, h.appState)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return false
		}
		writeError(w, http.StatusForbidden, "forbidden")
		return false
	}
	if !webaccess.CanLinkStreamerIntegrations(access) {
		writeError(w, http.StatusForbidden, "forbidden")
		return false
	}
	return true
}

func (h handler) requireBotIntegrationAccess(w http.ResponseWriter, r *http.Request) bool {
	if r == nil || h.appState == nil || h.appState.Sessions == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return false
	}
	sessionID := session.SessionIDFromRequest(r)
	if sessionID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return false
	}
	userSession, err := h.appState.Sessions.Get(r.Context(), sessionID)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return false
		}
		writeError(w, http.StatusForbidden, "forbidden")
		return false
	}
	access, err := webaccess.EvaluateDashboardAccess(r.Context(), h.appState, userSession.UserID, userSession.Login)
	if err != nil {
		writeError(w, http.StatusForbidden, "forbidden")
		return false
	}
	if !webaccess.CanLinkBotIntegration(access) {
		writeError(w, http.StatusForbidden, "forbidden")
		return false
	}
	return true
}
