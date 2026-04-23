package dashboard

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
)

type spamFilterResponse struct {
	ID                     string   `json:"id"`
	Name                   string   `json:"name"`
	Description            string   `json:"description"`
	Action                 string   `json:"action"`
	ThresholdLabel         string   `json:"threshold_label"`
	ThresholdValue         int      `json:"threshold_value"`
	Enabled                bool     `json:"enabled"`
	RepeatOffendersEnabled bool     `json:"repeat_offenders_enabled"`
	RepeatMultiplier       float64  `json:"repeat_multiplier"`
	RepeatMemorySeconds    int      `json:"repeat_memory_seconds"`
	RepeatUntilStreamEnd   bool     `json:"repeat_until_stream_end"`
	ImpactedRoles          []string `json:"impacted_roles"`
	ExcludedRoles          []string `json:"excluded_roles"`
}

type updateSpamFilterRequest struct {
	ID                     string    `json:"id"`
	Action                 string    `json:"action"`
	ThresholdLabel         string    `json:"threshold_label"`
	ThresholdValue         int       `json:"threshold_value"`
	Enabled                bool      `json:"enabled"`
	RepeatOffendersEnabled *bool     `json:"repeat_offenders_enabled"`
	RepeatMultiplier       *float64  `json:"repeat_multiplier"`
	RepeatMemorySeconds    *int      `json:"repeat_memory_seconds"`
	RepeatUntilStreamEnd   *bool     `json:"repeat_until_stream_end"`
	ImpactedRoles          *[]string `json:"impacted_roles"`
	ExcludedRoles          *[]string `json:"excluded_roles"`
}

func (h handler) spamFilters(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listSpamFilters(w, r)
	case http.MethodPut:
		h.updateSpamFilter(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet+", "+http.MethodPut)
	}
}

func (h handler) listSpamFilters(w http.ResponseWriter, r *http.Request) {
	if err := h.requireDashboardAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.SpamFilters == nil {
		http.Error(w, "spam filters are not configured", http.StatusInternalServerError)
		return
	}
	if err := h.appState.SpamFilters.EnsureDefaults(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	items, err := h.appState.SpamFilters.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]spamFilterResponse, 0, len(items))
	for _, item := range items {
		response = append(response, spamFilterToResponse(item))
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"filters": response})
}

func (h handler) updateSpamFilter(w http.ResponseWriter, r *http.Request) {
	if err := h.requireDashboardAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.SpamFilters == nil {
		http.Error(w, "spam filters are not configured", http.StatusInternalServerError)
		return
	}

	var request updateSpamFilterRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid spam filter payload", http.StatusBadRequest)
		return
	}

	current, err := h.appState.SpamFilters.Get(r.Context(), request.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if current == nil {
		http.Error(w, "spam filter not found", http.StatusNotFound)
		return
	}

	current.Action = strings.TrimSpace(request.Action)
	current.ThresholdLabel = strings.TrimSpace(request.ThresholdLabel)
	current.ThresholdValue = request.ThresholdValue
	current.Enabled = request.Enabled
	if request.RepeatOffendersEnabled != nil {
		current.RepeatOffendersEnabled = *request.RepeatOffendersEnabled
	}
	if request.RepeatMultiplier != nil {
		current.RepeatMultiplier = *request.RepeatMultiplier
	}
	if request.RepeatMemorySeconds != nil {
		current.RepeatMemorySeconds = *request.RepeatMemorySeconds
	}
	if request.RepeatUntilStreamEnd != nil {
		current.RepeatUntilStreamEnd = *request.RepeatUntilStreamEnd
	}
	if request.ImpactedRoles != nil {
		current.ImpactedRoles = append([]string(nil), (*request.ImpactedRoles)...)
	}
	if request.ExcludedRoles != nil {
		current.ExcludedRoles = append([]string(nil), (*request.ExcludedRoles)...)
	}

	updated, err := h.appState.SpamFilters.Update(r.Context(), *current)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(spamFilterToResponse(*updated))
}

func spamFilterToResponse(item postgres.SpamFilter) spamFilterResponse {
	return spamFilterResponse{
		ID:                     item.FilterKey,
		Name:                   item.Title,
		Description:            item.Description,
		Action:                 item.Action,
		ThresholdLabel:         item.ThresholdLabel,
		ThresholdValue:         item.ThresholdValue,
		Enabled:                item.Enabled,
		RepeatOffendersEnabled: item.RepeatOffendersEnabled,
		RepeatMultiplier:       item.RepeatMultiplier,
		RepeatMemorySeconds:    item.RepeatMemorySeconds,
		RepeatUntilStreamEnd:   item.RepeatUntilStreamEnd,
		ImpactedRoles:          append([]string(nil), item.ImpactedRoles...),
		ExcludedRoles:          append([]string(nil), item.ExcludedRoles...),
	}
}
