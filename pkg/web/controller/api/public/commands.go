package public

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/commands"
	"github.com/mr-cheeezz/dankbot/pkg/modules"
	defaultcommandsmodule "github.com/mr-cheeezz/dankbot/pkg/modules/defaultcommands"
	gamemodule "github.com/mr-cheeezz/dankbot/pkg/modules/game"
	modesmodule "github.com/mr-cheeezz/dankbot/pkg/modules/modes"
	spotifymodule "github.com/mr-cheeezz/dankbot/pkg/modules/now-playing"
	quotesmodule "github.com/mr-cheeezz/dankbot/pkg/modules/quotes"
	tabsmodule "github.com/mr-cheeezz/dankbot/pkg/modules/tabs"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
)

type commandsResponse struct {
	RegularItems   []commandGroupResponse `json:"regular_items"`
	ModeratorItems []commandGroupResponse `json:"moderator_items"`
}

type commandGroupResponse struct {
	Title    string                  `json:"title"`
	Commands []publicCommandResponse `json:"commands"`
}

type publicCommandResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Usage       string `json:"usage"`
	Example     string `json:"example"`
}

type publicCommandDoc struct {
	Audience    string
	Module      string
	Name        string
	Description string
	Usage       string
	Example     string
}

func (h handler) commands(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	response := commandsResponse{
		RegularItems:   h.buildCommandGroups(r.Context(), "regular"),
		ModeratorItems: h.buildCommandGroups(r.Context(), "moderator"),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (h handler) buildCommandGroups(ctx context.Context, audience string) []commandGroupResponse {
	definitions := publicCommandDefinitions()
	if len(definitions) == 0 {
		return nil
	}

	commandPrefix := "!"
	if h.appState != nil && h.appState.PublicHomeSettings != nil {
		if err := h.appState.PublicHomeSettings.EnsureDefault(ctx); err == nil {
			if settings, err := h.appState.PublicHomeSettings.Get(ctx); err == nil && settings != nil {
				commandPrefix = normalizePublicCommandPrefix(settings.CommandPrefix)
			}
		}
	}

	defaultSettings := map[string]postgres.DefaultCommandSetting{}
	if h.appState != nil && h.appState.Postgres != nil {
		store := postgres.NewDefaultCommandSettingStore(h.appState.Postgres)
		if settings, err := store.List(ctx); err == nil {
			for _, setting := range settings {
				name := strings.TrimSpace(strings.ToLower(setting.CommandName))
				if name != "" {
					defaultSettings[name] = setting
				}
			}
		}
	}

	grouped := make(map[string][]publicCommandResponse)
	for _, doc := range publicCommandDocs(definitions, defaultSettings, commandPrefix) {
		if doc.Audience != audience {
			continue
		}

		moduleName := publicCommandGroupLabel(doc.Module)
		grouped[moduleName] = append(grouped[moduleName], publicCommandResponse{
			Name:        strings.TrimSpace(doc.Name),
			Description: strings.TrimSpace(doc.Description),
			Usage:       strings.TrimSpace(doc.Usage),
			Example:     strings.TrimSpace(doc.Example),
		})
	}

	groupNames := make([]string, 0, len(grouped))
	for groupName := range grouped {
		groupNames = append(groupNames, groupName)
	}
	sort.Strings(groupNames)

	items := make([]commandGroupResponse, 0, len(groupNames))
	for _, groupName := range groupNames {
		commands := grouped[groupName]
		sort.SliceStable(commands, func(i, j int) bool {
			return commands[i].Name < commands[j].Name
		})
		items = append(items, commandGroupResponse{
			Title:    groupName,
			Commands: commands,
		})
	}

	return items
}

func publicCommandDocs(
	definitions []commands.Definition,
	defaultSettings map[string]postgres.DefaultCommandSetting,
	commandPrefix string,
) []publicCommandDoc {
	byName := make(map[string]commands.Definition, len(definitions))
	for _, definition := range definitions {
		name := strings.TrimSpace(strings.ToLower(definition.Name))
		if name == "" {
			continue
		}
		byName[name] = definition
	}

	items := make([]publicCommandDoc, 0, len(definitions))

	appendDefinition := func(name, audience string) {
		definition, ok := byName[strings.TrimSpace(strings.ToLower(name))]
		if !ok || !publicCommandEnabled(definition, defaultSettings) {
			return
		}

		items = append(items, publicCommandDoc{
			Audience:    audience,
			Module:      strings.TrimSpace(definition.Module),
			Name:        applyPublicCommandPrefix(strings.TrimSpace(definition.Name), commandPrefix),
			Description: strings.TrimSpace(definition.Description),
			Usage:       applyPublicCommandPrefix(strings.TrimSpace(definition.Usage), commandPrefix),
			Example:     applyPublicCommandPrefix(strings.TrimSpace(definition.Example), commandPrefix),
		})
	}

	appendCustom := func(baseName, audience, name, description, usage, example string) {
		definition, ok := byName[strings.TrimSpace(strings.ToLower(baseName))]
		if !ok || !publicCommandEnabled(definition, defaultSettings) {
			return
		}

		items = append(items, publicCommandDoc{
			Audience:    audience,
			Module:      strings.TrimSpace(definition.Module),
			Name:        applyPublicCommandPrefix(strings.TrimSpace(name), commandPrefix),
			Description: strings.TrimSpace(description),
			Usage:       applyPublicCommandPrefix(strings.TrimSpace(usage), commandPrefix),
			Example:     applyPublicCommandPrefix(strings.TrimSpace(example), commandPrefix),
		})
	}

	for _, name := range []string{"ping", "game", "playtime", "gamesplayed", "quote", "currentmode", "link"} {
		appendDefinition(name, "regular")
	}

	appendCustom(
		"song",
		"regular",
		"!song",
		"Shows the current Spotify song the streamer is listening to.",
		"!song",
		"!song",
	)
	appendCustom(
		"song",
		"regular",
		"!song next",
		"Shows the next song currently sitting in the Spotify queue.",
		"!song next",
		"!song next",
	)
	appendCustom(
		"song",
		"regular",
		"!song last",
		"Shows the last Spotify song that finished playing.",
		"!song last",
		"!song last",
	)

	for _, name := range []string{"mode", "modes", "killswitch", "ks", "add quote", "create quote", "del quote", "rm quote", "edit quote", "tab add", "tab set", "tab paid", "tab give"} {
		appendDefinition(name, "moderator")
	}

	appendDefinition("tab", "regular")

	appendCustom(
		"song",
		"moderator",
		"!song add",
		"Queues a Spotify song from a URL, URI, or search terms.",
		"!song add <spotify url|spotify uri|search terms>",
		"!song add https://open.spotify.com/track/...",
	)
	appendCustom(
		"song",
		"moderator",
		"!song skip",
		"Skips the currently playing Spotify song.",
		"!song skip",
		"!song skip",
	)
	appendCustom(
		"song",
		"moderator",
		"!song volume",
		"Sets Spotify playback volume for the active device.",
		"!song volume <0-100>",
		"!song volume 35",
	)

	return items
}

func publicCommandEnabled(definition commands.Definition, defaultSettings map[string]postgres.DefaultCommandSetting) bool {
	name := strings.TrimSpace(strings.ToLower(definition.Name))
	if name == "" {
		return false
	}

	if setting, ok := defaultSettings[name]; ok && !setting.Enabled {
		return false
	}

	return true
}

func publicCommandDefinitions() []commands.Definition {
	dispatcher := commands.NewDispatcher("!")
	runner := modules.NewRunner(dispatcher)

	runner.Register(defaultcommandsmodule.New(time.Now().UTC(), "dev", nil))
	runner.Register(spotifymodule.New(nil, nil, nil, nil))
	runner.Register(gamemodule.New("", "", "", nil, nil, nil, nil, nil))
	runner.Register(quotesmodule.New(nil, nil))
	runner.Register(modesmodule.New(nil, nil, nil, nil))
	runner.Register(tabsmodule.New(nil, nil))

	definitions := dispatcher.Definitions()
	sort.SliceStable(definitions, func(i, j int) bool {
		if definitions[i].Module == definitions[j].Module {
			return definitions[i].Name < definitions[j].Name
		}
		return definitions[i].Module < definitions[j].Module
	})

	return definitions
}

func publicCommandGroupLabel(moduleName string) string {
	switch strings.TrimSpace(moduleName) {
	case "default-commands":
		return "Default Commands"
	case "now-playing":
		return "Spotify Commands"
	case "game":
		return "Game Commands"
	case "quotes":
		return "Quote Commands"
	case "modes":
		return "Mode Commands"
	default:
		trimmed := strings.TrimSpace(moduleName)
		if trimmed == "" {
			return "Commands"
		}
		parts := strings.Split(strings.ReplaceAll(trimmed, "-", " "), " ")
		for index, part := range parts {
			if part == "" {
				continue
			}
			parts[index] = strings.ToUpper(part[:1]) + part[1:]
		}
		return strings.Join(parts, " ")
	}
}

func applyPublicCommandPrefix(value, prefix string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	prefix = normalizePublicCommandPrefix(prefix)
	if strings.HasPrefix(value, "!") {
		return prefix + strings.TrimPrefix(value, "!")
	}

	return prefix + value
}

func normalizePublicCommandPrefix(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "!"
	}

	return raw
}
