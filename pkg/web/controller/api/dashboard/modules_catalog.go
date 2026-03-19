package dashboard

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
)

type modulesCatalogResponse struct {
	Items []moduleCatalogEntryResponse `json:"items"`
}

type moduleCatalogEntryResponse struct {
	ID       string                       `json:"id"`
	Name     string                       `json:"name"`
	State    string                       `json:"state"`
	Detail   string                       `json:"detail"`
	Commands []string                     `json:"commands"`
	Settings []moduleCatalogSettingSchema `json:"settings"`
}

type moduleCatalogSettingSchema struct {
	ID         string   `json:"id"`
	Label      string   `json:"label"`
	Type       string   `json:"type"`
	HelperText string   `json:"helper_text,omitempty"`
	Options    []string `json:"options,omitempty"`
}

func (h handler) modulesCatalog(w http.ResponseWriter, r *http.Request) {
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

	if h.appState == nil || h.appState.ModuleCatalog == nil {
		http.Error(w, "module catalog is not configured", http.StatusInternalServerError)
		return
	}

	entries, err := h.appState.ModuleCatalog.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := modulesCatalogResponse{
		Items: make([]moduleCatalogEntryResponse, 0, len(entries)),
	}
	for _, entry := range entries {
		response.Items = append(response.Items, moduleCatalogEntryToResponse(entry))
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func moduleCatalogEntryToResponse(entry postgres.ModuleCatalogEntry) moduleCatalogEntryResponse {
	settings := make([]moduleCatalogSettingSchema, 0, len(entry.Settings))
	for _, setting := range entry.Settings {
		settings = append(settings, moduleCatalogSettingSchema{
			ID:         setting.ID,
			Label:      setting.Label,
			Type:       setting.Type,
			HelperText: setting.HelperText,
			Options:    append([]string(nil), setting.Options...),
		})
	}

	return moduleCatalogEntryResponse{
		ID:       entry.ID,
		Name:     entry.Name,
		State:    entry.State,
		Detail:   entry.Detail,
		Commands: append([]string(nil), entry.Commands...),
		Settings: settings,
	}
}
