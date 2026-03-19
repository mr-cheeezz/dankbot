package dashboard

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/mr-cheeezz/dankbot/pkg/web/session"
)

type twitchCategorySearchItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	BoxArtURL string `json:"box_art_url"`
}

type twitchCategorySearchResponse struct {
	Items []twitchCategorySearchItem `json:"items"`
}

func (h handler) twitchCategorySearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	if _, err := h.requireEditorFeatureAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(twitchCategorySearchResponse{Items: []twitchCategorySearchItem{}})
		return
	}

	client, err := h.dashboardAppHelixClient(r.Context())
	if err != nil {
		http.Error(w, err.Error(), moderationToolsStatusCode(err))
		return
	}
	if client == nil {
		http.Error(w, "twitch app client is unavailable", http.StatusPreconditionFailed)
		return
	}

	results, err := client.SearchCategories(r.Context(), query, 15)
	if err != nil {
		http.Error(w, err.Error(), moderationToolsStatusCode(err))
		return
	}

	items := make([]twitchCategorySearchItem, 0, len(results))
	for _, item := range results {
		id := strings.TrimSpace(item.ID)
		name := strings.TrimSpace(item.Name)
		if id == "" || name == "" {
			continue
		}
		items = append(items, twitchCategorySearchItem{
			ID:        id,
			Name:      name,
			BoxArtURL: strings.TrimSpace(item.BoxArtURL),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(twitchCategorySearchResponse{Items: items})
}
