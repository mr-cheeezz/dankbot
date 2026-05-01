package justlog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/modules"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
)

var twitchLoginPattern = regexp.MustCompile(`^[a-z0-9_]{2,25}$`)

type Module struct {
	enabledFallback bool
	baseURL         string
	apiKey          string
	configPath      string
	allowedIDs      map[string]struct{}
	settings        *postgres.RustLogModuleSettingsStore

	mu       sync.RWMutex
	channels map[string]struct{}
	rng      *rand.Rand
}

func New(
	enabled bool,
	baseURL string,
	apiKey string,
	configPath string,
	allowedIDs ...string,
) *Module {
	allowed := make(map[string]struct{}, len(allowedIDs))
	for _, id := range allowedIDs {
		id = strings.TrimSpace(id)
		if id != "" {
			allowed[id] = struct{}{}
		}
	}

	return &Module{
		enabledFallback: enabled,
		baseURL:         strings.TrimSpace(baseURL),
		apiKey:          strings.TrimSpace(apiKey),
		configPath:      strings.TrimSpace(configPath),
		allowedIDs:      allowed,
		channels:        make(map[string]struct{}),
		rng:             rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (m *Module) Name() string {
	return "rustlog"
}

func (m *Module) RegisterCommands() map[string]modules.CommandDefinition {
	return map[string]modules.CommandDefinition{
		"log": {
			Handler:     m.log,
			Description: "Shows RustLog status and managed channels.",
			Usage:       "!log [list]",
			Example:     "!log list",
		},
		"log status": {
			Handler:     m.log,
			Description: "Shows RustLog status and managed channels.",
			Usage:       "!log status",
			Example:     "!log status",
		},
		"log add": {
			Handler:     m.logAdd,
			Description: "Adds a channel to managed RustLog channels.",
			Usage:       "!log add <channel>",
			Example:     "!log add mr_cheeezz",
		},
		"log remove": {
			Handler:     m.logRemove,
			Description: "Removes a channel from managed RustLog channels.",
			Usage:       "!log remove <channel>",
			Example:     "!log remove mr_cheeezz",
		},
		"log rm": {
			Handler:     m.logRemove,
			Description: "Removes a channel from managed RustLog channels.",
			Usage:       "!log rm <channel>",
			Example:     "!log rm mr_cheeezz",
		},
		"log del": {
			Handler:     m.logRemove,
			Description: "Removes a channel from managed RustLog channels.",
			Usage:       "!log del <channel>",
			Example:     "!log del mr_cheeezz",
		},
	}
}

func (m *Module) Start(ctx context.Context) error {
	if m.settings != nil {
		if err := m.settings.EnsureDefault(ctx, m.enabledFallback); err != nil {
			return err
		}
	}
	_ = m.loadChannelsFromConfigFile()
	return nil
}

func (m *Module) SetSettingsStore(store *postgres.RustLogModuleSettingsStore) {
	m.settings = store
}

func (m *Module) log(ctx modules.CommandContext) (string, error) {
	if !m.isEnabled() {
		return "RustLog is disabled in config right now.", nil
	}

	if len(ctx.Args) > 0 && strings.EqualFold(strings.TrimSpace(ctx.Args[0]), "list") {
		channels := m.listChannels()
		if len(channels) == 0 {
			return "No RustLog channels are configured yet.", nil
		}
		return fmt.Sprintf("RustLog channels: %s", strings.Join(channels, ", ")), nil
	}

	channels := m.listChannels()
	if len(channels) == 0 {
		return m.pick([]string{
			"RustLog is online. No channels are managed yet. Use !log add <channel>.",
			"RustLog is enabled and waiting. Drop a channel with !log add <channel>.",
			"RustLog is ready. Start with !log add <channel> to track someone.",
		}), nil
	}

	return m.pick([]string{
		fmt.Sprintf("RustLog is enabled and tracking %d channel(s).", len(channels)),
		fmt.Sprintf("RustLog status: active, %d channel(s) managed.", len(channels)),
		fmt.Sprintf("Logging vibes are good. Currently managing %d channel(s).", len(channels)),
	}), nil
}

func (m *Module) logAdd(ctx modules.CommandContext) (string, error) {
	if !m.isEnabled() {
		return "RustLog is disabled in config right now.", nil
	}
	if !m.canManage(ctx) {
		return "Only moderators or the broadcaster can change RustLog channels.", nil
	}
	if len(ctx.Args) == 0 {
		return "usage: " + commandPrefix(ctx) + "log add <channel>", nil
	}

	channel, ok := normalizeChannel(ctx.Args[0])
	if !ok {
		return "That channel login looks invalid. Use a Twitch login like mr_cheeezz.", nil
	}

	m.mu.Lock()
	if _, exists := m.channels[channel]; exists {
		m.mu.Unlock()
		return m.pick([]string{
			fmt.Sprintf("%s is already in the RustLog list.", channel),
			fmt.Sprintf("%s was already being tracked by RustLog.", channel),
			fmt.Sprintf("No changes made. %s is already managed.", channel),
		}), nil
	}
	m.channels[channel] = struct{}{}
	m.mu.Unlock()

	warnings := m.syncExternal("add", channel)
	if len(warnings) > 0 {
		return fmt.Sprintf("Added %s, but some sync steps failed: %s", channel, strings.Join(warnings, " | ")), nil
	}

	return m.pick([]string{
		fmt.Sprintf("Added %s to RustLog and synced everything cleanly.", channel),
		fmt.Sprintf("%s is now managed and synced to your RustLog setup.", channel),
		fmt.Sprintf("Locked in. %s was added and all sync targets passed.", channel),
	}), nil
}

func (m *Module) logRemove(ctx modules.CommandContext) (string, error) {
	if !m.isEnabled() {
		return "RustLog is disabled in config right now.", nil
	}
	if !m.canManage(ctx) {
		return "Only moderators or the broadcaster can change RustLog channels.", nil
	}
	if len(ctx.Args) == 0 {
		return "usage: " + commandPrefix(ctx) + "log remove <channel>", nil
	}

	channel, ok := normalizeChannel(ctx.Args[0])
	if !ok {
		return "That channel login looks invalid. Use a Twitch login like mr_cheeezz.", nil
	}

	m.mu.Lock()
	if _, exists := m.channels[channel]; !exists {
		m.mu.Unlock()
		return m.pick([]string{
			fmt.Sprintf("%s is not in the RustLog list right now.", channel),
			fmt.Sprintf("No removal needed. %s was not being tracked.", channel),
			fmt.Sprintf("Couldn't find %s in managed channels.", channel),
		}), nil
	}
	delete(m.channels, channel)
	m.mu.Unlock()

	warnings := m.syncExternal("remove", channel)
	if len(warnings) > 0 {
		return fmt.Sprintf("Removed %s, but some sync steps failed: %s", channel, strings.Join(warnings, " | ")), nil
	}

	return m.pick([]string{
		fmt.Sprintf("Removed %s from RustLog and synced everything cleanly.", channel),
		fmt.Sprintf("%s is no longer managed and external sync is done.", channel),
		fmt.Sprintf("Done. %s was removed and all sync targets passed.", channel),
	}), nil
}

func (m *Module) syncExternal(action string, channel string) []string {
	warnings := make([]string, 0, 2)

	if err := m.syncAPI(action, channel); err != nil {
		warnings = append(warnings, "api: "+err.Error())
	}

	if err := m.syncConfigFile(); err != nil {
		warnings = append(warnings, "config file: "+err.Error())
	}

	return warnings
}

func (m *Module) syncAPI(action string, channel string) error {
	base := strings.TrimSpace(m.baseURL)
	if base == "" {
		return nil
	}

	httpClient := &http.Client{Timeout: 5 * time.Second}

	if action == "add" {
		payload := map[string]string{"channel": channel}
		body, _ := json.Marshal(payload)

		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, strings.TrimRight(base, "/")+"/channels", bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		if strings.TrimSpace(m.apiKey) != "" {
			req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(m.apiKey))
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
			return fmt.Errorf("status %d %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
		}
		return nil
	}

	if action == "remove" {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, strings.TrimRight(base, "/")+"/channels/"+channel, nil)
		if err != nil {
			return err
		}
		if strings.TrimSpace(m.apiKey) != "" {
			req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(m.apiKey))
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
			return fmt.Errorf("status %d %s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
		}
		return nil
	}

	return nil
}

func (m *Module) syncConfigFile() error {
	path := strings.TrimSpace(m.configPath)
	if path == "" {
		return nil
	}

	channels := m.listChannels()
	content := strings.Join(channels, "\n")
	if content != "" {
		content += "\n"
	}

	return os.WriteFile(path, []byte(content), 0o644)
}

func (m *Module) loadChannelsFromConfigFile() error {
	path := strings.TrimSpace(m.configPath)
	if path == "" {
		return nil
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	loaded := make(map[string]struct{})
	for _, line := range strings.Split(string(raw), "\n") {
		if normalized, ok := normalizeChannel(line); ok {
			loaded[normalized] = struct{}{}
		}
	}

	m.mu.Lock()
	m.channels = loaded
	m.mu.Unlock()
	return nil
}

func (m *Module) isEnabled() bool {
	if m.settings == nil {
		return m.enabledFallback
	}

	settings, err := m.settings.Get(context.Background())
	if err != nil || settings == nil {
		return m.enabledFallback
	}
	return settings.Enabled
}

func (m *Module) canManage(ctx modules.CommandContext) bool {
	if ctx.IsBroadcaster || ctx.IsModerator {
		return true
	}

	_, ok := m.allowedIDs[strings.TrimSpace(ctx.SenderID)]
	return ok
}

func (m *Module) listChannels() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	channels := make([]string, 0, len(m.channels))
	for channel := range m.channels {
		channels = append(channels, channel)
	}
	sort.Strings(channels)
	return channels
}

func (m *Module) pick(options []string) string {
	if len(options) == 0 {
		return ""
	}
	if len(options) == 1 {
		return options[0]
	}

	m.mu.Lock()
	index := m.rng.Intn(len(options))
	m.mu.Unlock()
	return options[index]
}

func normalizeChannel(value string) (string, bool) {
	value = strings.TrimSpace(strings.ToLower(strings.TrimPrefix(value, "@")))
	if !twitchLoginPattern.MatchString(value) {
		return "", false
	}
	return value, true
}

func commandPrefix(ctx modules.CommandContext) string {
	prefix := strings.TrimSpace(ctx.CommandPrefix)
	if prefix == "" {
		return "!"
	}
	return prefix
}
