package dashboard

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/commands"
	"github.com/mr-cheeezz/dankbot/pkg/modules"
	customcommandsmodule "github.com/mr-cheeezz/dankbot/pkg/modules/customcommands"
	defaultcommandsmodule "github.com/mr-cheeezz/dankbot/pkg/modules/defaultcommands"
	gamemodule "github.com/mr-cheeezz/dankbot/pkg/modules/game"
	modesmodule "github.com/mr-cheeezz/dankbot/pkg/modules/modes"
	spotifymodule "github.com/mr-cheeezz/dankbot/pkg/modules/now-playing"
	quotesmodule "github.com/mr-cheeezz/dankbot/pkg/modules/quotes"
	tabsmodule "github.com/mr-cheeezz/dankbot/pkg/modules/tabs"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
)

type commandListResponse struct {
	Items []commandResponse `json:"items"`
}

type commandResponse struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Kind               string   `json:"kind"`
	DefaultEnabled     bool     `json:"default_enabled"`
	Platform           string   `json:"platform"`
	Aliases            []string `json:"aliases"`
	Group              string   `json:"group"`
	State              string   `json:"state"`
	Description        string   `json:"description"`
	Example            string   `json:"example"`
	ResponsePreview    string   `json:"response_preview"`
	ResponseType       string   `json:"response_type"`
	Enabled            bool     `json:"enabled"`
	EnabledWhenOffline bool     `json:"enabled_when_offline"`
	EnabledWhenOnline  bool     `json:"enabled_when_online"`
	Protected          bool     `json:"protected"`
	Configurable       bool     `json:"configurable"`
}

func (h handler) commands(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listCommands(w, r)
	case http.MethodPost:
		h.createCommand(w, r)
	case http.MethodPut:
		h.updateCommand(w, r)
	case http.MethodDelete:
		h.deleteCommand(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet+", "+http.MethodPost+", "+http.MethodPut+", "+http.MethodDelete)
	}
}

func (h handler) listCommands(w http.ResponseWriter, r *http.Request) {
	if err := h.requireDashboardAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	items, err := h.buildDashboardCommands(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(commandListResponse{Items: items})
}

func (h handler) createCommand(w http.ResponseWriter, r *http.Request) {
	userSession, err := h.requireEditorFeatureAccess(r)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if h.appState == nil || h.appState.CustomCommands == nil {
		http.Error(w, "custom command storage is not configured", http.StatusInternalServerError)
		return
	}

	var request commandResponse
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid command payload", http.StatusBadRequest)
		return
	}
	entry, err := validateCustomCommandPayload(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	entry.ID = fmt.Sprintf("cmd-custom-%d", time.Now().UTC().UnixNano())
	entry.UpdatedBy = strings.TrimSpace(userSession.Login)

	existing, err := h.appState.CustomCommands.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := validateCustomCommandConflicts(entry, existing, nil); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	created, err := h.appState.CustomCommands.Create(r.Context(), entry)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(customCommandToResponse(*created))
}

func (h handler) updateCommand(w http.ResponseWriter, r *http.Request) {
	userSession, err := h.requireEditorFeatureAccess(r)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var request commandResponse
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid command payload", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(strings.ToLower(request.Kind)) == "default" {
		if h.appState == nil || h.appState.Postgres == nil {
			http.Error(w, "default command storage is not configured", http.StatusInternalServerError)
			return
		}
		store := postgres.NewDefaultCommandSettingStore(h.appState.Postgres)
		if err := store.SetEnabled(r.Context(), request.Name, request.Enabled, strings.TrimSpace(userSession.Login)); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		items, err := h.buildDashboardCommands(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, item := range items {
			if strings.TrimSpace(item.Kind) == "default" &&
				strings.EqualFold(strings.TrimSpace(item.Name), strings.TrimSpace(request.Name)) &&
				strings.EqualFold(strings.TrimSpace(item.Platform), strings.TrimSpace(request.Platform)) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(item)
				return
			}
		}
		http.Error(w, "default command not found", http.StatusNotFound)
		return
	}

	if h.appState == nil || h.appState.CustomCommands == nil {
		http.Error(w, "custom command storage is not configured", http.StatusInternalServerError)
		return
	}

	entry, err := validateCustomCommandPayload(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	entry.ID = strings.TrimSpace(request.ID)
	entry.UpdatedBy = strings.TrimSpace(userSession.Login)

	existing, err := h.appState.CustomCommands.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := validateCustomCommandConflicts(entry, existing, &entry.ID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updated, err := h.appState.CustomCommands.Update(r.Context(), entry)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if updated == nil {
		http.Error(w, "custom command not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(customCommandToResponse(*updated))
}

func (h handler) deleteCommand(w http.ResponseWriter, r *http.Request) {
	if _, err := h.requireEditorFeatureAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if h.appState == nil || h.appState.CustomCommands == nil {
		http.Error(w, "custom command storage is not configured", http.StatusInternalServerError)
		return
	}

	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		http.Error(w, "command id is required", http.StatusBadRequest)
		return
	}
	deleted, err := h.appState.CustomCommands.Delete(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !deleted {
		http.Error(w, "custom command not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h handler) buildDashboardCommands(ctx context.Context) ([]commandResponse, error) {
	defaults, err := h.buildDefaultDashboardCommands(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]commandResponse, 0, len(defaults)+16)
	items = append(items, defaults...)

	if h.appState != nil && h.appState.CustomCommands != nil {
		customs, err := h.appState.CustomCommands.List(ctx)
		if err != nil {
			return nil, err
		}
		for _, entry := range customs {
			items = append(items, customCommandToResponse(entry))
		}
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Platform == items[j].Platform {
			if items[i].Kind == items[j].Kind {
				return items[i].Name < items[j].Name
			}
			return items[i].Kind < items[j].Kind
		}
		return items[i].Platform < items[j].Platform
	})
	return items, nil
}

func (h handler) buildDefaultDashboardCommands(ctx context.Context) ([]commandResponse, error) {
	dispatcher := commands.NewDispatcher("!")
	runner := modules.NewRunner(dispatcher)
	runner.Register(customcommandsmodule.New(nil))
	runner.Register(defaultcommandsmodule.New(time.Now().UTC(), "dev", nil))
	runner.Register(spotifymodule.New(nil, nil, nil, nil))
	runner.Register(gamemodule.New("", "", "", nil, nil, nil, nil, nil))
	runner.Register(quotesmodule.New(nil, nil))
	runner.Register(modesmodule.New(nil, nil, nil, nil))
	runner.Register(tabsmodule.New(nil, nil))

	definitions := dispatcher.Definitions()
	defaultSettings := map[string]postgres.DefaultCommandSetting{}
	if h.appState != nil && h.appState.Postgres != nil {
		store := postgres.NewDefaultCommandSettingStore(h.appState.Postgres)
		settings, err := store.List(ctx)
		if err != nil {
			return nil, err
		}
		for _, setting := range settings {
			defaultSettings[strings.TrimSpace(strings.ToLower(setting.CommandName))] = setting
		}
	}

	items := make([]commandResponse, 0, len(definitions))
	for _, definition := range definitions {
		name := strings.TrimSpace(definition.Name)
		if name == "" {
			continue
		}
		enabled := definition.DefaultEnabled
		if setting, ok := defaultSettings[strings.ToLower(name)]; ok {
			enabled = setting.Enabled
		}
		items = append(items, commandResponse{
			ID:                 "default:" + name,
			Name:               name,
			Kind:               "default",
			DefaultEnabled:     definition.DefaultEnabled,
			Platform:           "twitch",
			Aliases:            []string{},
			Group:              strings.TrimSpace(definition.Module),
			State:              defaultCommandState(enabled, definition),
			Description:        strings.TrimSpace(definition.Description),
			Example:            strings.TrimSpace(definition.Example),
			ResponsePreview:    "",
			ResponseType:       "reply",
			Enabled:            enabled,
			EnabledWhenOffline: true,
			EnabledWhenOnline:  true,
			Protected:          !definition.CanDisable,
			Configurable:       definition.Configurable,
		})
	}
	return items, nil
}

func defaultCommandState(enabled bool, definition commands.Definition) string {
	if !definition.CanDisable {
		return "always on"
	}
	if enabled {
		return "enabled"
	}
	return "disabled"
}

func validateCustomCommandPayload(request commandResponse) (postgres.CustomCommand, error) {
	entry := postgres.CustomCommand{
		Platform:           strings.TrimSpace(request.Platform),
		Name:               strings.TrimSpace(request.Name),
		Aliases:            append([]string(nil), request.Aliases...),
		Group:              strings.TrimSpace(request.Group),
		State:              strings.TrimSpace(request.State),
		Description:        strings.TrimSpace(request.Description),
		Example:            strings.TrimSpace(request.Example),
		ResponseTemplate:   strings.TrimSpace(request.ResponsePreview),
		ResponseType:       strings.TrimSpace(request.ResponseType),
		Enabled:            request.Enabled,
		EnabledWhenOffline: request.EnabledWhenOffline,
		EnabledWhenOnline:  request.EnabledWhenOnline,
	}
	entry = postgres.CustomCommand{
		ID:                 strings.TrimSpace(request.ID),
		Platform:           strings.ToLower(strings.TrimSpace(entry.Platform)),
		Name:               strings.ToLower(strings.TrimSpace(entry.Name)),
		Aliases:            normalizeDashboardCommandAliases(entry.Aliases),
		Group:              strings.TrimSpace(entry.Group),
		State:              strings.TrimSpace(entry.State),
		Description:        strings.TrimSpace(entry.Description),
		Example:            strings.TrimSpace(entry.Example),
		ResponseTemplate:   strings.TrimSpace(entry.ResponseTemplate),
		ResponseType:       strings.ToLower(strings.TrimSpace(entry.ResponseType)),
		Enabled:            entry.Enabled,
		EnabledWhenOffline: entry.EnabledWhenOffline,
		EnabledWhenOnline:  entry.EnabledWhenOnline,
	}
	if entry.Platform == "" {
		entry.Platform = "twitch"
	}
	if entry.Platform != "twitch" && entry.Platform != "discord" {
		return postgres.CustomCommand{}, fmt.Errorf("unsupported command platform")
	}
	if entry.Name == "" {
		return postgres.CustomCommand{}, fmt.Errorf("command name is required")
	}
	if entry.ResponseTemplate == "" {
		return postgres.CustomCommand{}, fmt.Errorf("command response is required")
	}
	switch entry.ResponseType {
	case "", "reply", "say", "action":
		if entry.ResponseType == "" {
			entry.ResponseType = "reply"
		}
	default:
		return postgres.CustomCommand{}, fmt.Errorf("unsupported command response type")
	}
	if entry.Group == "" {
		if entry.Platform == "discord" {
			entry.Group = "discord"
		} else {
			entry.Group = "custom"
		}
	}
	if entry.State == "" {
		if entry.Enabled {
			entry.State = "enabled"
		} else {
			entry.State = "disabled"
		}
	}
	return entry, nil
}

func normalizeDashboardCommandAliases(aliases []string) []string {
	normalized := make([]string, 0, len(aliases))
	seen := make(map[string]struct{}, len(aliases))
	for _, alias := range aliases {
		value := strings.ToLower(strings.TrimSpace(alias))
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	return normalized
}

func validateCustomCommandConflicts(entry postgres.CustomCommand, existing []postgres.CustomCommand, ignoreID *string) error {
	ignore := ""
	if ignoreID != nil {
		ignore = strings.TrimSpace(*ignoreID)
	}
	targets := append([]string{entry.Name}, entry.Aliases...)
	for _, current := range existing {
		if ignore != "" && strings.TrimSpace(current.ID) == ignore {
			continue
		}
		if current.Platform != entry.Platform {
			continue
		}
		currentNames := append([]string{current.Name}, current.Aliases...)
		for _, target := range targets {
			for _, currentName := range currentNames {
				if strings.EqualFold(strings.TrimSpace(target), strings.TrimSpace(currentName)) {
					return fmt.Errorf("command name or alias %q is already in use", target)
				}
			}
		}
	}
	return nil
}

func customCommandToResponse(entry postgres.CustomCommand) commandResponse {
	return commandResponse{
		ID:                 entry.ID,
		Name:               entry.Name,
		Kind:               "custom",
		DefaultEnabled:     false,
		Platform:           entry.Platform,
		Aliases:            append([]string(nil), entry.Aliases...),
		Group:              entry.Group,
		State:              entry.State,
		Description:        entry.Description,
		Example:            entry.Example,
		ResponsePreview:    entry.ResponseTemplate,
		ResponseType:       entry.ResponseType,
		Enabled:            entry.Enabled,
		EnabledWhenOffline: entry.EnabledWhenOffline,
		EnabledWhenOnline:  entry.EnabledWhenOnline,
		Protected:          false,
		Configurable:       true,
	}
}
