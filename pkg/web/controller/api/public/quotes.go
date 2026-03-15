package public

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mr-cheeezz/dankbot/pkg/postgres"
)

type quotesResponse struct {
	Items []quoteResponse `json:"items"`
}

type quoteResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h handler) quotes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	response := quotesResponse{
		Items: []quoteResponse{},
	}

	if h.appState == nil || h.appState.Postgres == nil {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
		return
	}

	store := postgres.NewQuoteStore(h.appState.Postgres)
	items, err := store.List(r.Context(), 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response.Items = make([]quoteResponse, 0, len(items))
	for _, item := range items {
		response.Items = append(response.Items, quoteResponse{
			ID:      item.ID,
			Message: strings.TrimSpace(item.Message),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
