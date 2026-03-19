package dashboard

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
)

type quoteModuleResponse struct {
	Enabled bool `json:"enabled"`
}

type quoteEntryResponse struct {
	ID        int64  `json:"id"`
	Message   string `json:"message"`
	CreatedBy string `json:"created_by"`
	UpdatedBy string `json:"updated_by"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type quoteEntriesResponse struct {
	Items []quoteEntryResponse `json:"items"`
}

type quoteEntryRequest struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

func (h handler) quoteModule(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getQuoteModule(w, r)
	case http.MethodPut:
		h.updateQuoteModule(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet+", "+http.MethodPut)
	}
}

func (h handler) quoteModuleEntries(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listQuoteEntries(w, r)
	case http.MethodPost:
		h.createQuoteEntry(w, r)
	case http.MethodPut:
		h.updateQuoteEntry(w, r)
	case http.MethodDelete:
		h.deleteQuoteEntry(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet+", "+http.MethodPost+", "+http.MethodPut+", "+http.MethodDelete)
	}
}

func (h handler) getQuoteModule(w http.ResponseWriter, r *http.Request) {
	if _, err := h.requireEditorFeatureAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.QuoteModule == nil {
		http.Error(w, "quote module settings are not configured", http.StatusInternalServerError)
		return
	}

	if err := h.appState.QuoteModule.EnsureDefault(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	settings, err := h.appState.QuoteModule.Get(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if settings == nil {
		defaults := postgres.DefaultQuoteModuleSettings()
		settings = &defaults
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(quoteModuleResponse{
		Enabled: settings.Enabled,
	})
}

func (h handler) updateQuoteModule(w http.ResponseWriter, r *http.Request) {
	userSession, err := h.requireEditorFeatureAccess(r)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.QuoteModule == nil {
		http.Error(w, "quote module settings are not configured", http.StatusInternalServerError)
		return
	}

	if err := h.appState.QuoteModule.EnsureDefault(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var request quoteModuleResponse
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid quote module payload", http.StatusBadRequest)
		return
	}

	updated, err := h.appState.QuoteModule.Update(r.Context(), postgres.QuoteModuleSettings{
		Enabled:   request.Enabled,
		UpdatedBy: strings.TrimSpace(userSession.Login),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if updated == nil {
		http.Error(w, "quote module settings not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(quoteModuleResponse{
		Enabled: updated.Enabled,
	})
}

func (h handler) listQuoteEntries(w http.ResponseWriter, r *http.Request) {
	if _, err := h.requireEditorFeatureAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.Postgres == nil {
		http.Error(w, "quote storage is not configured", http.StatusInternalServerError)
		return
	}

	quotes, err := postgres.NewQuoteStore(h.appState.Postgres).List(r.Context(), 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := quoteEntriesResponse{
		Items: make([]quoteEntryResponse, 0, len(quotes)),
	}
	for _, quote := range quotes {
		response.Items = append(response.Items, quoteToResponse(quote))
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (h handler) createQuoteEntry(w http.ResponseWriter, r *http.Request) {
	userSession, err := h.requireEditorFeatureAccess(r)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.Postgres == nil {
		http.Error(w, "quote storage is not configured", http.StatusInternalServerError)
		return
	}

	var request quoteEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid quote payload", http.StatusBadRequest)
		return
	}

	created, err := postgres.NewQuoteStore(h.appState.Postgres).Create(
		r.Context(),
		strings.TrimSpace(request.Message),
		strings.TrimSpace(userSession.Login),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(quoteToResponse(*created))
}

func (h handler) updateQuoteEntry(w http.ResponseWriter, r *http.Request) {
	userSession, err := h.requireEditorFeatureAccess(r)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.Postgres == nil {
		http.Error(w, "quote storage is not configured", http.StatusInternalServerError)
		return
	}

	var request quoteEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid quote payload", http.StatusBadRequest)
		return
	}
	if request.ID <= 0 {
		http.Error(w, "quote id is required", http.StatusBadRequest)
		return
	}

	updated, err := postgres.NewQuoteStore(h.appState.Postgres).Update(
		r.Context(),
		request.ID,
		strings.TrimSpace(request.Message),
		strings.TrimSpace(userSession.Login),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if updated == nil {
		http.Error(w, "quote not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(quoteToResponse(*updated))
}

func (h handler) deleteQuoteEntry(w http.ResponseWriter, r *http.Request) {
	if _, err := h.requireEditorFeatureAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.Postgres == nil {
		http.Error(w, "quote storage is not configured", http.StatusInternalServerError)
		return
	}

	var request quoteEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid quote payload", http.StatusBadRequest)
		return
	}
	if request.ID <= 0 {
		http.Error(w, "quote id is required", http.StatusBadRequest)
		return
	}

	deleted, err := postgres.NewQuoteStore(h.appState.Postgres).Delete(r.Context(), request.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !deleted {
		http.Error(w, "quote not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"id":     strconv.FormatInt(request.ID, 10),
	})
}

func quoteToResponse(quote postgres.Quote) quoteEntryResponse {
	return quoteEntryResponse{
		ID:        quote.ID,
		Message:   strings.TrimSpace(quote.Message),
		CreatedBy: strings.TrimSpace(quote.CreatedBy),
		UpdatedBy: strings.TrimSpace(quote.UpdatedBy),
		CreatedAt: quote.CreatedAt.Format(time.RFC3339),
		UpdatedAt: quote.UpdatedAt.Format(time.RFC3339),
	}
}
