package dashboard

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
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

type quoteImportRequest struct {
	Source   string `json:"source"`
	Payload  string `json:"payload"`
	Channel  string `json:"channel"`
	APIURL   string `json:"api_url"`
	APIToken string `json:"api_token"`
}

type quoteImportResponse struct {
	Imported int                  `json:"imported"`
	Skipped  int                  `json:"skipped"`
	Items    []quoteEntryResponse `json:"items"`
}

var quoteLinePrefixPattern = regexp.MustCompile(`^(?:[#\[]?\d+[\]\)]?\s*[:.)\-]?\s*)(.+)$`)

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

func (h handler) quoteModuleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}

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

	var request quoteImportRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid quote import payload", http.StatusBadRequest)
		return
	}

	source := strings.ToLower(strings.TrimSpace(request.Source))
	if source == "" {
		source = "fossabot"
	}
	if source != "fossabot" {
		http.Error(w, "unsupported quote import source", http.StatusBadRequest)
		return
	}

	parsedMessages := parseFossabotQuotes(request.Payload)
	if len(parsedMessages) == 0 {
		var fetchErr error
		parsedMessages, fetchErr = fetchFossabotQuotesFromAPI(r.Context(), request)
		if fetchErr != nil {
			http.Error(w, fetchErr.Error(), http.StatusBadRequest)
			return
		}
	}
	if len(parsedMessages) == 0 {
		http.Error(w, "no quotes found to import", http.StatusBadRequest)
		return
	}

	store := postgres.NewQuoteStore(h.appState.Postgres)
	existing, err := store.List(r.Context(), 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	seen := make(map[string]struct{}, len(existing)+len(parsedMessages))
	for _, item := range existing {
		key := normalizeQuoteImportMessage(item.Message)
		if key == "" {
			continue
		}
		seen[key] = struct{}{}
	}

	created := make([]quoteEntryResponse, 0, len(parsedMessages))
	skipped := 0
	actor := strings.TrimSpace(userSession.Login)

	for _, message := range parsedMessages {
		key := normalizeQuoteImportMessage(message)
		if key == "" {
			skipped++
			continue
		}
		if _, exists := seen[key]; exists {
			skipped++
			continue
		}

		item, createErr := store.Create(r.Context(), message, actor)
		if createErr != nil {
			http.Error(w, createErr.Error(), http.StatusBadRequest)
			return
		}
		seen[key] = struct{}{}
		created = append(created, quoteToResponse(*item))
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(quoteImportResponse{
		Imported: len(created),
		Skipped:  skipped,
		Items:    created,
	})
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

func parseFossabotQuotes(raw string) []string {
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	raw = strings.ReplaceAll(raw, "\r", "\n")
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	type jsonQuote struct {
		Quote   string `json:"quote"`
		Message string `json:"message"`
		Text    string `json:"text"`
	}
	var jsonItems []jsonQuote
	if err := json.Unmarshal([]byte(raw), &jsonItems); err == nil && len(jsonItems) > 0 {
		out := make([]string, 0, len(jsonItems))
		for _, item := range jsonItems {
			message := strings.TrimSpace(item.Quote)
			if message == "" {
				message = strings.TrimSpace(item.Message)
			}
			if message == "" {
				message = strings.TrimSpace(item.Text)
			}
			if message != "" {
				out = append(out, message)
			}
		}
		return out
	}

	lines := strings.Split(raw, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		value := strings.TrimSpace(line)
		if value == "" {
			continue
		}
		value = strings.TrimPrefix(value, "-")
		value = strings.TrimPrefix(value, "*")
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		if match := quoteLinePrefixPattern.FindStringSubmatch(value); len(match) >= 2 {
			value = strings.TrimSpace(match[1])
		}

		if index := strings.Index(value, "\t"); index > 0 {
			left := strings.TrimSpace(value[:index])
			right := strings.TrimSpace(value[index+1:])
			if left != "" && right != "" {
				if _, err := strconv.Atoi(strings.TrimLeft(left, "#")); err == nil {
					value = right
				}
			}
		}

		if value != "" {
			out = append(out, value)
		}
	}

	return out
}

func normalizeQuoteImportMessage(message string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(message)), " "))
}

func fetchFossabotQuotesFromAPI(ctx context.Context, request quoteImportRequest) ([]string, error) {
	channel := strings.TrimSpace(strings.ToLower(request.Channel))
	apiURL := strings.TrimSpace(request.APIURL)
	apiToken := strings.TrimSpace(request.APIToken)

	candidates := make([]string, 0, 4)
	if apiURL != "" {
		candidates = append(candidates, apiURL)
	}
	if channel != "" {
		escaped := url.PathEscape(channel)
		candidates = append(candidates,
			"https://api.fossabot.com/v2/channels/"+escaped+"/quotes",
			"https://api.fossabot.com/v1/channels/"+escaped+"/quotes",
			"https://fossabot.com/v2/channels/"+escaped+"/quotes",
		)
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("provide Fossabot channel or API URL")
	}

	client := &http.Client{Timeout: 15 * time.Second}
	lastErr := ""
	for _, endpoint := range candidates {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			lastErr = err.Error()
			continue
		}
		req.Header.Set("Accept", "application/json, text/plain;q=0.8, */*;q=0.5")
		if apiToken != "" {
			req.Header.Set("Authorization", "Bearer "+apiToken)
			req.Header.Set("X-API-Key", apiToken)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err.Error()
			continue
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 4*1024*1024))
		_ = resp.Body.Close()
		if readErr != nil {
			lastErr = readErr.Error()
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastErr = strings.TrimSpace(fmt.Sprintf("%d %s", resp.StatusCode, string(body)))
			continue
		}

		quotes := parseFossabotQuotes(string(body))
		if len(quotes) == 0 {
			quotes = parseFossabotQuotesJSON(body)
		}
		if len(quotes) > 0 {
			return quotes, nil
		}
		lastErr = "response did not contain quote entries"
	}

	if lastErr == "" {
		lastErr = "unknown error"
	}
	return nil, fmt.Errorf("could not fetch Fossabot quotes via API: %s", lastErr)
}

func parseFossabotQuotesJSON(raw []byte) []string {
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil
	}

	collections := []any{
		data["quotes"],
		data["items"],
		data["data"],
	}
	for _, collection := range collections {
		items, ok := collection.([]any)
		if !ok || len(items) == 0 {
			continue
		}
		out := make([]string, 0, len(items))
		for _, item := range items {
			object, ok := item.(map[string]any)
			if !ok {
				continue
			}
			candidates := []string{
				stringValue(object["quote"]),
				stringValue(object["message"]),
				stringValue(object["text"]),
				stringValue(object["content"]),
			}
			for _, candidate := range candidates {
				if strings.TrimSpace(candidate) != "" {
					out = append(out, strings.TrimSpace(candidate))
					break
				}
			}
		}
		if len(out) > 0 {
			return out
		}
	}

	return nil
}

func stringValue(value any) string {
	text, _ := value.(string)
	return strings.TrimSpace(text)
}
