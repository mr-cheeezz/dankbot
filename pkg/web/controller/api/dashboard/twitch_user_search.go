package dashboard

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	webaccess "github.com/mr-cheeezz/dankbot/pkg/web/access"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
)

type twitchUserSearchItem struct {
	UserID      string `json:"user_id"`
	Login       string `json:"login"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

type twitchUserSearchResponse struct {
	Items []twitchUserSearchItem `json:"items"`
}

func (h handler) twitchUserSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

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

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(twitchUserSearchResponse{Items: []twitchUserSearchItem{}})
		return
	}

	users, err := webaccess.SearchTwitchUsers(r.Context(), h.appState, query, 8)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	items := make([]twitchUserSearchItem, 0, len(users))
	for _, user := range users {
		items = append(items, twitchUserSearchItem{
			UserID:      user.ID,
			Login:       user.Login,
			DisplayName: user.DisplayName,
			AvatarURL:   user.ProfileImageURL,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(twitchUserSearchResponse{Items: items})
}
