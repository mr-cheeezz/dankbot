package sessionauth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	webaccess "github.com/mr-cheeezz/dankbot/pkg/web/access"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
	"github.com/mr-cheeezz/dankbot/pkg/web/state"
)

type handler struct {
	appState *state.State
}

type sessionUserResponse struct {
	UserID             string `json:"user_id"`
	Login              string `json:"login"`
	DisplayName        string `json:"display_name"`
	AvatarURL          string `json:"avatar_url"`
	IsModerator        bool   `json:"is_moderator"`
	IsVIP              bool   `json:"is_vip"`
	IsLeadModerator    bool   `json:"is_lead_moderator"`
	IsBroadcaster      bool   `json:"is_broadcaster"`
	IsBotAccount       bool   `json:"is_bot_account"`
	IsEditor           bool   `json:"is_editor"`
	IsAdmin            bool   `json:"is_admin"`
	CanAccessDashboard bool   `json:"can_access_dashboard"`
}

type sessionResponse struct {
	LoggedIn           bool                 `json:"logged_in"`
	CanAccessDashboard bool                 `json:"can_access_dashboard"`
	User               *sessionUserResponse `json:"user,omitempty"`
}

func Register(mux *http.ServeMux, appState *state.State) {
	h := handler{appState: appState}
	mux.Handle("/api/auth/session", http.HandlerFunc(h.status))
	mux.Handle("/auth/logout", http.HandlerFunc(h.logout))
}

func (h handler) status(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	response := sessionResponse{
		LoggedIn:           false,
		CanAccessDashboard: false,
	}

	userSession, ok := h.lookupSession(r)
	if ok {
		userSession = h.refreshAccess(r, sessionIDFromRequest(r), userSession)
		response.LoggedIn = true
		response.CanAccessDashboard = userSession.CanAccessDashboard
		response.User = &sessionUserResponse{
			UserID:             userSession.UserID,
			Login:              userSession.Login,
			DisplayName:        userSession.DisplayName,
			AvatarURL:          userSession.AvatarURL,
			IsModerator:        userSession.IsModerator,
			IsVIP:              userSession.IsVIP,
			IsLeadModerator:    userSession.IsLeadModerator,
			IsBroadcaster:      userSession.IsBroadcaster,
			IsBotAccount:       userSession.IsBotAccount,
			IsEditor:           userSession.IsEditor,
			IsAdmin:            userSession.IsAdmin,
			CanAccessDashboard: userSession.CanAccessDashboard,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (h handler) logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}

	sessionID := session.SessionIDFromRequest(r)
	if sessionID != "" && h.appState != nil && h.appState.Sessions != nil {
		_ = h.appState.Sessions.Delete(r.Context(), sessionID)
	}

	session.ClearCookie(w, isSecureCookie(h.appState))
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (h handler) lookupSession(r *http.Request) (*session.UserSession, bool) {
	if r == nil || h.appState == nil || h.appState.Sessions == nil {
		return nil, false
	}

	sessionID := session.SessionIDFromRequest(r)
	if sessionID == "" {
		return nil, false
	}

	userSession, err := h.appState.Sessions.Get(r.Context(), sessionID)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			return nil, false
		}
		return nil, false
	}

	return userSession, true
}

func (h handler) refreshAccess(r *http.Request, sessionID string, userSession *session.UserSession) *session.UserSession {
	if userSession == nil || h.appState == nil {
		return userSession
	}

	access, err := webaccess.EvaluateDashboardAccess(r.Context(), h.appState, userSession.UserID, userSession.Login)
	if err != nil {
		return userSession
	}

	next := *userSession
	next.IsModerator = access.IsModerator
	next.IsVIP = access.IsVIP
	next.IsLeadModerator = access.IsLeadModerator
	next.IsBroadcaster = access.IsBroadcaster
	next.IsBotAccount = access.IsBotAccount
	next.IsEditor = access.IsEditor
	next.IsAdmin = access.IsAdmin
	next.CanAccessDashboard = access.CanAccessDashboard

	if h.appState.Sessions != nil && sessionID != "" {
		_ = h.appState.Sessions.Save(r.Context(), sessionID, next)
	}

	return &next
}

func sessionIDFromRequest(r *http.Request) string {
	return session.SessionIDFromRequest(r)
}

func isSecureCookie(appState *state.State) bool {
	if appState == nil || appState.Config == nil {
		return false
	}

	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(appState.Config.Web.PublicURL)), "https://")
}

func writeMethodNotAllowed(w http.ResponseWriter, allowedMethod string) {
	w.Header().Set("Allow", allowedMethod)
	w.WriteHeader(http.StatusMethodNotAllowed)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
}
