package dashboard

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	webaccess "github.com/mr-cheeezz/dankbot/pkg/web/access"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
)

type dashboardRoleResponse struct {
	UserID          string `json:"user_id"`
	Login           string `json:"login"`
	DisplayName     string `json:"display_name"`
	RoleName        string `json:"role_name"`
	AssignedByLogin string `json:"assigned_by_login"`
}

type dashboardRolesResponse struct {
	Items []dashboardRoleResponse `json:"items"`
}

type assignEditorRequest struct {
	UserID      string `json:"user_id"`
	Login       string `json:"login"`
	DisplayName string `json:"display_name"`
}

type deleteEditorRequest struct {
	UserID string `json:"user_id"`
}

var twitchLoginPattern = regexp.MustCompile(`(?i)[a-z0-9_]{2,25}`)

func (h handler) roles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getRoles(w, r)
	case http.MethodPost:
		h.assignEditorRole(w, r)
	case http.MethodDelete:
		h.deleteEditorRole(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet+", "+http.MethodPost+", "+http.MethodDelete)
	}
}

func (h handler) getRoles(w http.ResponseWriter, r *http.Request) {
	userSession, err := h.dashboardSession(r)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if !canManageDashboardRoles(userSession) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	items, err := h.listDashboardRoles(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dashboardRolesResponse{Items: items})
}

func (h handler) assignEditorRole(w http.ResponseWriter, r *http.Request) {
	userSession, err := h.dashboardSession(r)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if !canManageDashboardRoles(userSession) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if h.appState == nil || h.appState.DashboardRoles == nil {
		http.Error(w, "dashboard roles are not configured", http.StatusInternalServerError)
		return
	}

	var request assignEditorRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid dashboard role payload", http.StatusBadRequest)
		return
	}

	userID := strings.TrimSpace(request.UserID)
	login := normalizeTwitchLogin(request.Login)
	if login == "" {
		http.Error(w, "login is required", http.StatusBadRequest)
		return
	}

	displayName := strings.TrimSpace(request.DisplayName)
	if userID == "" {
		user, err := webaccess.LookupTwitchUserByLogin(r.Context(), h.appState, login)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if user == nil {
			http.Error(w, "twitch user not found", http.StatusNotFound)
			return
		}
		userID = strings.TrimSpace(user.ID)
		login = strings.TrimSpace(user.Login)
		displayName = strings.TrimSpace(user.DisplayName)
	}
	if userID == "" {
		http.Error(w, "user id is required", http.StatusBadRequest)
		return
	}
	if displayName == "" {
		displayName = login
	}

	if err := h.appState.DashboardRoles.Save(r.Context(), postgres.DashboardRole{
		UserID:          userID,
		Login:           login,
		DisplayName:     displayName,
		RoleName:        postgres.DashboardRoleEditor,
		AssignedByLogin: strings.TrimSpace(userSession.Login),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	items, err := h.listDashboardRoles(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dashboardRolesResponse{Items: items})
}

func (h handler) deleteEditorRole(w http.ResponseWriter, r *http.Request) {
	userSession, err := h.dashboardSession(r)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if !canManageDashboardRoles(userSession) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if h.appState == nil || h.appState.DashboardRoles == nil {
		http.Error(w, "dashboard roles are not configured", http.StatusInternalServerError)
		return
	}

	var request deleteEditorRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid dashboard role payload", http.StatusBadRequest)
		return
	}

	userID := strings.TrimSpace(request.UserID)
	if userID == "" {
		http.Error(w, "user id is required", http.StatusBadRequest)
		return
	}

	if err := h.appState.DashboardRoles.Delete(r.Context(), userID, postgres.DashboardRoleEditor); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	items, err := h.listDashboardRoles(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dashboardRolesResponse{Items: items})
}

func (h handler) listDashboardRoles(r *http.Request) ([]dashboardRoleResponse, error) {
	if h.appState == nil || h.appState.DashboardRoles == nil {
		return nil, nil
	}

	items, err := h.appState.DashboardRoles.List(r.Context())
	if err != nil {
		return nil, err
	}

	out := make([]dashboardRoleResponse, 0, len(items))
	for _, item := range items {
		out = append(out, dashboardRoleResponse{
			UserID:          item.UserID,
			Login:           item.Login,
			DisplayName:     item.DisplayName,
			RoleName:        string(item.RoleName),
			AssignedByLogin: item.AssignedByLogin,
		})
	}

	return out, nil
}

func canManageDashboardRoles(userSession *session.UserSession) bool {
	if userSession == nil {
		return false
	}

	return userSession.IsBroadcaster || userSession.IsAdmin
}

func normalizeTwitchLogin(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}

	value = strings.ToLower(value)
	if match := twitchLoginPattern.FindString(value); match != "" {
		return match
	}
	return ""
}
