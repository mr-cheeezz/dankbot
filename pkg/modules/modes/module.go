package modes

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/modules"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	twitchhelix "github.com/mr-cheeezz/dankbot/pkg/twitch/helix"
	twitchoauth "github.com/mr-cheeezz/dankbot/pkg/twitch/oauth"
)

const (
	modeTitleEnforceInterval = 15 * time.Second
	modeTitleEnforceTimeout  = 5 * time.Second
)

var (
	robloxURLPattern = regexp.MustCompile(`https?://[^\s<>()]+`)
)

var builtInModes = []postgres.BotMode{
	{
		ModeKey:              "join",
		Title:                "Join",
		Description:          "Default join mode.",
		KeywordName:          "join",
		KeywordDescription:   "Viewer-facing response for people asking how they can join while join mode is active.",
		KeywordResponse:      "@{target}, {streamer} is currently taking join requests. Use the active join keyword shown in chat once to get in.",
		IsBuiltin:            true,
		TimerEnabled:         true,
		TimerMessage:         "Type !join to play!",
		TimerIntervalSeconds: 240,
	},
	{
		ModeKey:              "link",
		Title:                "Link",
		Description:          "Posts or references the active Roblox private server link flow.",
		KeywordName:          "link",
		KeywordDescription:   "Viewer-facing response for people asking how to join during link mode.",
		KeywordResponse:      "@{target}, {streamer} is currently using link mode. Use the posted link or the website join panel to get in.",
		IsBuiltin:            true,
		TimerEnabled:         true,
		TimerMessage:         "Type !link to join!",
		TimerIntervalSeconds: 240,
	},
	{
		ModeKey:              "1v1",
		Title:                "1v1",
		Description:          "Game-specific 1v1 mode.",
		KeywordName:          "1v1",
		KeywordDescription:   "Viewer-facing response for people asking whether 1v1s are happening.",
		KeywordResponse:      "@{target}, type 1v1 in the chat ONCE to have a chance to 1v1 {streamer}.",
		IsBuiltin:            true,
		TimerEnabled:         true,
		TimerMessage:         "Type 1v1 for a chance to 1v1 {streamer}!",
		TimerIntervalSeconds: 180,
	},
	{
		ModeKey:              "reddit",
		Title:                "Reddit",
		Description:          "Reddit recap mode.",
		KeywordName:          "reddit",
		KeywordDescription:   "Viewer-facing response for people asking about the subreddit or recap link.",
		KeywordResponse:      "@{target}, {streamer} is currently using reddit mode. Use the active reddit command or website prompt for the link.",
		IsBuiltin:            true,
		TimerEnabled:         true,
		TimerMessage:         "Type !reddit for a link to the subreddit!",
		TimerIntervalSeconds: 180,
	},
}

func BuiltInModes() []postgres.BotMode {
	out := make([]postgres.BotMode, len(builtInModes))
	copy(out, builtInModes)
	return out
}

type Module struct {
	modeStore       *postgres.BotModeStore
	stateStore      *postgres.BotStateStore
	socialStore     *postgres.BotSocialPromotionStore
	auditStore      *postgres.AuditLogStore
	channelSettings *postgres.PublicHomeSettingsStore
	settingsStore   *postgres.ModesModuleSettingsStore
	adminIDs        map[string]struct{}
	isLive          func(context.Context) (bool, error)
	twitchOAuth     *twitchoauth.Service
	accounts        *postgres.TwitchAccountStore
	clientID        string

	mu      sync.RWMutex
	channel string
	say     func(channel, message string) error
}

func New(modeStore *postgres.BotModeStore, stateStore *postgres.BotStateStore, socialStore *postgres.BotSocialPromotionStore, auditStore *postgres.AuditLogStore, allowedIDs ...string) *Module {
	adminIDs := make(map[string]struct{})
	for _, id := range allowedIDs {
		id = strings.TrimSpace(id)
		if id != "" {
			adminIDs[id] = struct{}{}
		}
	}

	return &Module{
		modeStore:   modeStore,
		stateStore:  stateStore,
		socialStore: socialStore,
		auditStore:  auditStore,
		adminIDs:    adminIDs,
	}
}

func (m *Module) SetTwitchTitleCoordinator(clientID string, oauth *twitchoauth.Service, accounts *postgres.TwitchAccountStore) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.clientID = strings.TrimSpace(clientID)
	m.twitchOAuth = oauth
	m.accounts = accounts
}

func (m *Module) SetChannelSettingsStore(store *postgres.PublicHomeSettingsStore) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.channelSettings = store
}

func (m *Module) SetModesModuleSettingsStore(store *postgres.ModesModuleSettingsStore) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.settingsStore = store
}

func (m *Module) Name() string {
	return "modes"
}

func (m *Module) RegisterCommands() map[string]modules.CommandDefinition {
	return map[string]modules.CommandDefinition{
		"mode": {
			Handler:     m.mode,
			Description: "Switches the bot into one of the configured modes.",
			Usage:       "!mode <name> [param]",
			Example:     "!mode join",
		},
		"modes": {
			Handler:     m.modes,
			Description: "Lists the valid bot modes.",
			Usage:       "!modes",
			Example:     "!modes",
		},
		"currentmode": {
			Handler:     m.currentMode,
			Description: "Shows the mode the bot is currently running.",
			Usage:       "!currentmode",
			Example:     "!currentmode",
		},
		"killswitch": {
			Handler:     m.killswitch,
			Description: "Toggles the bot killswitch on or off.",
			Usage:       "!killswitch",
			Example:     "!killswitch",
		},
		"ks": {
			Handler:     m.killswitch,
			Description: "Alias for !killswitch.",
			Usage:       "!ks",
			Example:     "!ks",
		},
		"link": {
			Handler:     m.link,
			Description: "Shows the current Roblox private server link when link mode is active.",
			Usage:       "!link",
			Example:     "!link",
		},
	}
}

func (m *Module) Start(ctx context.Context) error {
	if m.modeStore == nil || m.stateStore == nil {
		return fmt.Errorf("modes stores are not configured")
	}

	if err := m.ensureBuiltInModes(ctx); err != nil {
		return err
	}

	if err := m.stateStore.Ensure(ctx, "join"); err != nil {
		return err
	}
	if m.settingsStore != nil {
		if err := m.settingsStore.EnsureDefault(ctx); err != nil {
			return err
		}
	}

	go m.runModeTimer(ctx)
	go m.runSocialTimer(ctx)
	go m.runTitleEnforcer(ctx)

	return nil
}

func (m *Module) HandleMessage(ctx modules.CommandContext) (modules.MessageResult, error) {
	trimmedMessage := strings.TrimSpace(ctx.Message)
	commandPrefix := normalizeCommandPrefix(ctx.CommandPrefix)
	reply, handledLegacy, err := m.handleLegacyModeCommand(ctx, trimmedMessage, commandPrefix)
	if err != nil {
		return modules.MessageResult{}, err
	}
	if handledLegacy {
		return modules.MessageResult{
			Reply:          strings.TrimSpace(reply),
			StopProcessing: true,
		}, nil
	}
	if strings.HasPrefix(trimmedMessage, commandPrefix) {
		return modules.MessageResult{}, nil
	}
	if m.isLikelyBotSender(ctx) {
		return modules.MessageResult{}, nil
	}
	if !m.canManageModes(ctx) {
		return modules.MessageResult{}, nil
	}

	link := detectRobloxPrivateServerLink(ctx.Message)
	if link == "" {
		return modules.MessageResult{}, nil
	}

	state, err := m.stateStore.Get(context.Background())
	if err != nil {
		return modules.MessageResult{}, err
	}
	if state != nil &&
		strings.EqualFold(strings.TrimSpace(state.CurrentModeKey), "link") &&
		sameRobloxPrivateServerLink(state.CurrentModeParam, link) {
		// Ignore duplicate link reposts to prevent noisy mode re-announcements.
		return modules.MessageResult{}, nil
	}

	if err := m.stateStore.SetCurrentMode(context.Background(), "link", link, ctx.SenderID); err != nil {
		return modules.MessageResult{}, err
	}
	warning, err := m.syncConfiguredLinkCommand(context.Background(), link)
	if err != nil {
		return modules.MessageResult{}, err
	}

	m.logAction(ctx, fmt.Sprintf("detected Roblox private server link and enabled link mode for %s", strings.ToLower(link)))

	return modules.MessageResult{
		Reply:          strings.TrimSpace(fmt.Sprintf("Link mode enabled. %slink now points to the posted private server. %s", commandPrefix, warning)),
		StopProcessing: true,
	}, nil
}

func (m *Module) mode(ctx modules.CommandContext) (string, error) {
	if !m.canManageModes(ctx) {
		return "", nil
	}
	if len(ctx.Args) == 0 {
		return m.currentMode(ctx)
	}

	modeKey := strings.TrimSpace(strings.ToLower(ctx.Args[0]))
	modeParam := ""
	if len(ctx.Args) > 1 {
		modeParam = strings.TrimSpace(strings.Join(ctx.Args[1:], " "))
	}

	return m.activateMode(ctx, modeKey, modeParam)
}

func (m *Module) activateMode(ctx modules.CommandContext, modeKey, modeParam string) (string, error) {
	modeKey = strings.TrimSpace(strings.ToLower(modeKey))
	modeParam = strings.TrimSpace(modeParam)
	if modeKey == "" {
		return "Mode key is required.", nil
	}

	mode, err := m.modeStore.Get(context.Background(), modeKey)
	if err != nil {
		return "", err
	}
	if mode == nil {
		return fmt.Sprintf(`Unknown mode "%s".`, modeKey), nil
	}

	state, err := m.stateStore.Get(context.Background())
	if err != nil {
		return "", err
	}
	if state != nil && state.CurrentModeKey == mode.ModeKey && strings.EqualFold(strings.TrimSpace(state.CurrentModeParam), strings.TrimSpace(modeParam)) {
		if strings.EqualFold(mode.ModeKey, "link") && sameRobloxPrivateServerLink(state.CurrentModeParam, modeParam) {
			// Keep duplicate link mode activations silent so reposting the same private
			// server link does not generate extra chat noise.
			return "", nil
		}
		warning := ""
		if strings.EqualFold(mode.ModeKey, "link") && looksLikeRobloxPrivateServerURL(modeParam) {
			warning, err = m.syncConfiguredLinkCommand(context.Background(), modeParam)
			if err != nil {
				return "", err
			}
		}
		if modeParam == "" {
			return strings.TrimSpace(fmt.Sprintf("%s mode is already active. %s", mode.ModeKey, warning)), nil
		}
		return strings.TrimSpace(fmt.Sprintf("%s mode is already active for %s. %s", mode.ModeKey, strings.ToLower(modeParam), warning)), nil
	}

	if err := m.stateStore.SetCurrentMode(context.Background(), mode.ModeKey, modeParam, ctx.SenderID); err != nil {
		return "", err
	}
	warning := ""
	if strings.EqualFold(mode.ModeKey, "link") && looksLikeRobloxPrivateServerURL(modeParam) {
		warning, err = m.syncConfiguredLinkCommand(context.Background(), modeParam)
		if err != nil {
			return "", err
		}
	}
	m.logAction(ctx, formatModeAuditDetail(mode.ModeKey, modeParam))

	streamerLogin := m.streamerLogin()
	senderName := m.senderName(ctx)
	if modeParam == "" {
		return strings.TrimSpace(fmt.Sprintf("@%s, %s has turned on %s mode. %s", streamerLogin, senderName, mode.ModeKey, warning)), nil
	}

	return strings.TrimSpace(fmt.Sprintf("@%s, %s has turned on %s mode for %s. %s", streamerLogin, senderName, mode.ModeKey, strings.ToLower(modeParam), warning)), nil
}

func (m *Module) handleLegacyModeCommand(ctx modules.CommandContext, message, commandPrefix string) (string, bool, error) {
	if !strings.HasPrefix(message, commandPrefix) {
		return "", false, nil
	}

	commandBody := strings.TrimSpace(strings.TrimPrefix(message, commandPrefix))
	if commandBody == "" {
		return "", false, nil
	}

	parts := strings.Fields(commandBody)
	if len(parts) != 1 {
		return "", false, nil
	}

	token := strings.TrimSpace(strings.ToLower(parts[0]))
	if !strings.HasSuffix(token, ".on") {
		return "", false, nil
	}

	enabled, err := m.legacyCommandsEnabled(context.Background())
	if err != nil {
		return "", false, err
	}
	if !enabled {
		return "", true, nil
	}

	if m.isLikelyBotSender(ctx) || !m.canManageModes(ctx) {
		return "", true, nil
	}

	modeKey := strings.TrimSuffix(token, ".on")
	if modeKey == "" {
		return "", true, nil
	}

	reply, err := m.activateMode(ctx, modeKey, "")
	return reply, true, err
}

func (m *Module) legacyCommandsEnabled(ctx context.Context) (bool, error) {
	m.mu.RLock()
	store := m.settingsStore
	m.mu.RUnlock()
	if store == nil {
		return false, nil
	}

	if err := store.EnsureDefault(ctx); err != nil {
		return false, err
	}

	settings, err := store.Get(ctx)
	if err != nil {
		return false, err
	}
	if settings == nil {
		defaults := postgres.DefaultModesModuleSettings()
		return defaults.LegacyCommandsEnabled, nil
	}

	return settings.LegacyCommandsEnabled, nil
}

func (m *Module) link(ctx modules.CommandContext) (string, error) {
	state, err := m.stateStore.Get(context.Background())
	if err != nil {
		return "", err
	}
	if state == nil {
		return "There is no active private server link right now.", nil
	}

	modeKey := strings.TrimSpace(strings.ToLower(state.CurrentModeKey))
	modeParam := strings.TrimSpace(state.CurrentModeParam)
	if modeKey != "link" || modeParam == "" || !looksLikeRobloxPrivateServerURL(modeParam) {
		return "There is no active private server link right now.", nil
	}

	return modeParam, nil
}

func (m *Module) modes(ctx modules.CommandContext) (string, error) {
	_ = ctx

	items, err := m.modeStore.List(context.Background())
	if err != nil {
		return "", err
	}
	if len(items) == 0 {
		return "there are no configured modes.", nil
	}

	names := make([]string, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.ModeKey) != "" {
			names = append(names, item.ModeKey)
		}
	}

	if len(names) == 0 {
		return "there are no configured modes.", nil
	}

	return "valid modes: " + strings.Join(names, ", "), nil
}

func (m *Module) currentMode(ctx modules.CommandContext) (string, error) {
	_ = ctx

	state, err := m.stateStore.Get(context.Background())
	if err != nil {
		return "", err
	}
	if state == nil {
		return "the bot is currently in join mode.", nil
	}

	mode, err := m.modeStore.Get(context.Background(), state.CurrentModeKey)
	if err != nil {
		return "", err
	}

	name := strings.TrimSpace(state.CurrentModeKey)
	if mode != nil && strings.TrimSpace(mode.ModeKey) != "" {
		name = mode.ModeKey
	}
	if name == "" {
		name = "unknown"
	}
	name = strings.ToLower(name)

	if strings.TrimSpace(state.CurrentModeParam) == "" {
		return fmt.Sprintf("the bot is currently in %s mode.", name), nil
	}

	return fmt.Sprintf("the bot is currently in %s mode for %s.", name, strings.ToLower(strings.TrimSpace(state.CurrentModeParam))), nil
}

func (m *Module) killswitch(ctx modules.CommandContext) (string, error) {
	if !m.canManageModes(ctx) {
		return "", nil
	}

	state, err := m.stateStore.ToggleKillswitch(context.Background(), ctx.SenderID)
	if err != nil {
		return "", err
	}
	if state != nil && state.KillswitchEnabled {
		m.logAction(ctx, "turned killswitch on")
		return "Killswitch enabled.", nil
	}

	m.logAction(ctx, "turned killswitch off")
	return "Killswitch disabled.", nil
}

func (m *Module) canManageModes(ctx modules.CommandContext) bool {
	if ctx.IsBroadcaster || ctx.IsModerator {
		return true
	}

	return m.isAllowedID(ctx.SenderID)
}

func (m *Module) isAllowedID(senderID string) bool {
	senderID = strings.TrimSpace(senderID)
	if senderID == "" {
		return false
	}

	_, ok := m.adminIDs[senderID]
	return ok
}

func (m *Module) SetChatOutput(channel string, say func(channel, message string) error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.channel = strings.TrimSpace(channel)
	m.say = say
}

func (m *Module) SetStreamLiveChecker(checker func(context.Context) (bool, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.isLive = checker
}

func (m *Module) runModeTimer(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := m.tickModeTimer(ctx); err != nil {
				fmt.Printf("mode timer error: %v\n", err)
			}
		}
	}
}

func (m *Module) runSocialTimer(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := m.tickSocialTimer(ctx); err != nil {
				fmt.Printf("social timer error: %v\n", err)
			}
		}
	}
}

func (m *Module) runTitleEnforcer(ctx context.Context) {
	ticker := time.NewTicker(modeTitleEnforceInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := m.tickTitleEnforcer(ctx); err != nil {
				fmt.Printf("mode title enforcement error: %v\n", err)
			}
		}
	}
}

func (m *Module) tickModeTimer(ctx context.Context) error {
	channel, say := m.output()
	if channel == "" || say == nil {
		return nil
	}

	state, err := m.stateStore.Get(ctx)
	if err != nil {
		return err
	}
	if state == nil || state.KillswitchEnabled {
		return nil
	}
	if m.isLive != nil {
		live, err := m.isLive(ctx)
		if err != nil {
			return err
		}
		if !live {
			return nil
		}
	}

	mode, err := m.modeStore.Get(ctx, state.CurrentModeKey)
	if err != nil {
		return err
	}
	if mode == nil || !mode.TimerEnabled || strings.TrimSpace(mode.TimerMessage) == "" || mode.TimerIntervalSeconds <= 0 {
		return nil
	}

	now := time.Now().UTC()
	if !mode.LastTimerSentAt.IsZero() && now.Sub(mode.LastTimerSentAt) < time.Duration(mode.TimerIntervalSeconds)*time.Second {
		return nil
	}

	message := strings.TrimSpace(mode.TimerMessage)
	if strings.TrimSpace(state.CurrentModeParam) != "" {
		message = strings.ReplaceAll(message, "{mode_param}", state.CurrentModeParam)
	}

	if err := say(channel, message); err != nil {
		return err
	}

	return m.modeStore.MarkTimerSent(ctx, mode.ModeKey, now)
}

func (m *Module) tickSocialTimer(ctx context.Context) error {
	channel, say := m.output()
	if channel == "" || say == nil || m.socialStore == nil {
		return nil
	}

	state, err := m.stateStore.Get(ctx)
	if err != nil {
		return err
	}
	if state != nil && state.KillswitchEnabled {
		return nil
	}
	if m.isLive != nil {
		live, err := m.isLive(ctx)
		if err != nil {
			return err
		}
		if !live {
			return nil
		}
	}

	items, err := m.socialStore.ListEnabled(ctx)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	for _, item := range items {
		if item.IntervalSeconds <= 0 || strings.TrimSpace(item.CommandText) == "" {
			continue
		}
		if !item.LastSentAt.IsZero() && now.Sub(item.LastSentAt) < time.Duration(item.IntervalSeconds)*time.Second {
			continue
		}

		if err := say(channel, item.CommandText); err != nil {
			return err
		}
		if err := m.socialStore.MarkSent(ctx, item.ID, now); err != nil {
			return err
		}
	}

	return nil
}

func (m *Module) tickTitleEnforcer(ctx context.Context) error {
	if m.modeStore == nil || m.stateStore == nil {
		return nil
	}

	clientID, oauthService, accountStore := m.titleCoordinator()
	if clientID == "" || oauthService == nil || accountStore == nil {
		return nil
	}

	state, err := m.stateStore.Get(ctx)
	if err != nil {
		return err
	}
	if state == nil || state.KillswitchEnabled {
		return nil
	}

	mode, err := m.modeStore.Get(ctx, state.CurrentModeKey)
	if err != nil {
		return err
	}
	if mode == nil {
		return nil
	}

	desiredTitle := strings.TrimSpace(mode.CoordinatedTwitchTitle)
	desiredCategoryID := strings.TrimSpace(mode.CoordinatedTwitchCategoryID)
	if desiredTitle == "" && desiredCategoryID == "" {
		return nil
	}

	account, err := m.streamerAccountForTitleSync(ctx, oauthService, accountStore)
	if err != nil {
		return err
	}
	if account == nil || strings.TrimSpace(account.TwitchUserID) == "" {
		return nil
	}
	if len(modeMissingScopes(account.Scopes, "channel:manage:broadcast")) > 0 {
		return nil
	}

	requestCtx, cancel := context.WithTimeout(ctx, modeTitleEnforceTimeout)
	defer cancel()

	client := twitchhelix.NewClientWithHTTPClient(
		&http.Client{Timeout: modeTitleEnforceTimeout},
		clientID,
		account.AccessToken,
	)

	channels, err := client.GetChannelsByBroadcasterIDs(requestCtx, []string{account.TwitchUserID})
	if err != nil {
		return err
	}
	if len(channels) == 0 {
		return nil
	}

	currentTitle := strings.TrimSpace(channels[0].Title)
	currentCategoryID := strings.TrimSpace(channels[0].GameID)

	request := twitchhelix.UpdateChannelInformationRequest{}
	if desiredTitle != "" && currentTitle != desiredTitle {
		request.Title = &desiredTitle
	}
	if desiredCategoryID != "" && currentCategoryID != desiredCategoryID {
		request.GameID = &desiredCategoryID
	}

	if request.Title == nil && request.GameID == nil {
		return nil
	}

	return client.UpdateChannelInformation(requestCtx, account.TwitchUserID, request)
}

func (m *Module) output() (string, func(channel, message string) error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.channel, m.say
}

func (m *Module) titleCoordinator() (string, *twitchoauth.Service, *postgres.TwitchAccountStore) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.clientID, m.twitchOAuth, m.accounts
}

func (m *Module) streamerLogin() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	channel := strings.TrimSpace(m.channel)
	if channel == "" {
		return "streamer"
	}

	return channel
}

func (m *Module) senderName(ctx modules.CommandContext) string {
	if name := strings.TrimSpace(ctx.DisplayName); name != "" {
		return name
	}
	if name := strings.TrimSpace(ctx.Sender); name != "" {
		return name
	}

	return "someone"
}

func (m *Module) isLikelyBotSender(ctx modules.CommandContext) bool {
	if ctx.IsBroadcaster {
		return false
	}

	knownBots := map[string]struct{}{
		"dankbot":        {},
		"nightbot":       {},
		"streamelements": {},
		"streamlabs":     {},
		"moobot":         {},
		"fossabot":       {},
		"pajbot":         {},
		"wizebot":        {},
		"phantombot":     {},
		"deepbot":        {},
	}

	candidates := []string{
		strings.ToLower(strings.TrimSpace(ctx.Sender)),
		strings.ToLower(strings.TrimSpace(ctx.DisplayName)),
	}

	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if _, ok := knownBots[candidate]; ok {
			return true
		}
		if strings.HasSuffix(candidate, "bot") {
			return true
		}
	}

	return false
}

func (m *Module) logAction(ctx modules.CommandContext, detail string) {
	if m.auditStore == nil || strings.TrimSpace(detail) == "" {
		return
	}

	command := strings.TrimSpace(ctx.Command)
	if command == "" {
		command = "mode"
	}
	command = normalizeCommandPrefix(ctx.CommandPrefix) + strings.TrimPrefix(command, "!")

	if _, err := m.auditStore.Create(context.Background(), postgres.AuditLog{
		Platform:  ctx.Platform,
		ActorID:   strings.TrimSpace(ctx.SenderID),
		ActorName: m.senderName(ctx),
		Command:   command,
		Detail:    detail,
	}); err != nil {
		fmt.Printf("audit log error: %v\n", err)
	}
}

func normalizeCommandPrefix(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "!"
	}

	return trimmed
}

func formatModeAuditDetail(modeKey, modeParam string) string {
	modeKey = strings.TrimSpace(strings.ToLower(modeKey))
	modeParam = strings.TrimSpace(modeParam)

	if modeParam == "" {
		return fmt.Sprintf("turned on %s mode", modeKey)
	}

	return fmt.Sprintf("turned on %s mode for %s", modeKey, strings.ToLower(modeParam))
}

func detectRobloxPrivateServerLink(message string) string {
	candidates := robloxURLPattern.FindAllString(message, -1)
	for _, candidate := range candidates {
		candidate = sanitizeDetectedURL(candidate)
		if looksLikeRobloxPrivateServerURL(candidate) {
			return candidate
		}
	}

	return ""
}

func sanitizeDetectedURL(raw string) string {
	return strings.TrimRight(strings.TrimSpace(raw), ".,!?;:)]}>\"'")
}

func looksLikeRobloxPrivateServerURL(raw string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}
	if !strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https") {
		return false
	}

	host := strings.ToLower(strings.TrimSpace(parsed.Hostname()))
	if host == "" || (host != "roblox.com" && !strings.HasSuffix(host, ".roblox.com")) {
		return false
	}

	query := parsed.Query()
	if strings.TrimSpace(query.Get("privateServerLinkCode")) != "" {
		return true
	}

	if strings.TrimSpace(query.Get("code")) == "" {
		return false
	}

	switch strings.ToLower(strings.TrimSpace(query.Get("type"))) {
	case "server", "privateserver":
		return true
	default:
		return false
	}
}

func sameRobloxPrivateServerLink(left, right string) bool {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	if left == "" || right == "" {
		return false
	}

	leftCode := robloxPrivateServerCode(left)
	rightCode := robloxPrivateServerCode(right)
	if leftCode != "" && rightCode != "" {
		return strings.EqualFold(leftCode, rightCode)
	}

	// Fallback to a strict normalized URL compare if code extraction fails.
	return strings.EqualFold(left, right)
}

func robloxPrivateServerCode(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return ""
	}

	query := parsed.Query()
	if code := strings.TrimSpace(query.Get("privateServerLinkCode")); code != "" {
		return code
	}
	if code := strings.TrimSpace(query.Get("code")); code != "" {
		if kind := strings.ToLower(strings.TrimSpace(query.Get("type"))); kind == "server" || kind == "privateserver" {
			return code
		}
	}

	return ""
}

func (m *Module) syncConfiguredLinkCommand(ctx context.Context, link string) (string, error) {
	target, template := m.linkCommandSettings(ctx)
	if target == "dankbot" {
		return "", nil
	}
	if strings.TrimSpace(template) == "" {
		return "No external link command template is configured yet.", nil
	}

	channel, say := m.output()
	if channel == "" || say == nil {
		return "The external link command could not be synced right now.", nil
	}

	command := strings.TrimSpace(strings.ReplaceAll(template, "{link}", strings.TrimSpace(link)))
	if command == "" {
		return "The external link command template is empty.", nil
	}

	if err := say(channel, command); err != nil {
		return "I could not sync the external !link command right now.", nil
	}

	return fmt.Sprintf("%s link command synced.", displayLinkCommandTarget(target)), nil
}

func (m *Module) linkCommandSettings(ctx context.Context) (string, string) {
	m.mu.RLock()
	store := m.channelSettings
	m.mu.RUnlock()

	if store == nil {
		return "dankbot", ""
	}

	if err := store.EnsureDefault(ctx); err != nil {
		return "dankbot", ""
	}

	settings, err := store.Get(ctx)
	if err != nil || settings == nil {
		return "dankbot", ""
	}

	target := strings.ToLower(strings.TrimSpace(settings.RobloxLinkCommandTarget))
	if target == "" {
		target = "dankbot"
	}

	return target, strings.TrimSpace(settings.RobloxLinkCommandTemplate)
}

func displayLinkCommandTarget(target string) string {
	switch strings.ToLower(strings.TrimSpace(target)) {
	case "nightbot":
		return "Nightbot"
	case "fossabot":
		return "Fossabot"
	case "pajbot":
		return "Pajbot"
	case "custom":
		return "Custom bot"
	default:
		return "DankBot"
	}
}

func (m *Module) ensureBuiltInModes(ctx context.Context) error {
	if err := m.modeStore.EnsureDefaults(ctx, BuiltInModes()); err != nil {
		return err
	}

	for _, builtInMode := range builtInModes {
		current, err := m.modeStore.Get(ctx, builtInMode.ModeKey)
		if err != nil {
			return err
		}
		if current == nil {
			continue
		}

		next := *current
		changed := false
		if !next.IsBuiltin {
			next.IsBuiltin = true
			changed = true
		}
		if strings.TrimSpace(next.Title) == "" {
			next.Title = builtInMode.Title
			changed = true
		}
		if strings.TrimSpace(next.Description) == "" {
			next.Description = builtInMode.Description
			changed = true
		}
		if strings.TrimSpace(next.KeywordName) == "" {
			next.KeywordName = builtInMode.KeywordName
			changed = true
		}
		if strings.TrimSpace(next.KeywordDescription) == "" {
			next.KeywordDescription = builtInMode.KeywordDescription
			changed = true
		}
		if strings.TrimSpace(next.KeywordResponse) == "" {
			next.KeywordResponse = builtInMode.KeywordResponse
			changed = true
		}
		if strings.TrimSpace(next.TimerMessage) == "" {
			next.TimerMessage = builtInMode.TimerMessage
			changed = true
		}
		if next.TimerIntervalSeconds <= 0 {
			next.TimerIntervalSeconds = builtInMode.TimerIntervalSeconds
			changed = true
		}

		if changed {
			if err := m.modeStore.Save(ctx, next); err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Module) streamerAccountForTitleSync(ctx context.Context, oauthService *twitchoauth.Service, accountStore *postgres.TwitchAccountStore) (*postgres.TwitchAccount, error) {
	account, err := accountStore.Get(ctx, postgres.TwitchAccountKindStreamer)
	if err != nil || account == nil {
		return account, err
	}

	if oauthService == nil || strings.TrimSpace(account.RefreshToken) == "" || !modeTokenNeedsRefresh(account.AccessToken, account.ExpiresAt) {
		return account, nil
	}

	token, err := oauthService.RefreshToken(ctx, account.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("refresh twitch streamer token for mode title sync: %w", err)
	}

	validation, err := oauthService.ValidateToken(ctx, token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("validate refreshed twitch streamer token for mode title sync: %w", err)
	}

	account.AccessToken = strings.TrimSpace(token.AccessToken)
	if refreshToken := strings.TrimSpace(token.RefreshToken); refreshToken != "" {
		account.RefreshToken = refreshToken
	}
	if len(token.Scope) > 0 {
		account.Scopes = append([]string(nil), token.Scope...)
	}
	if tokenType := strings.TrimSpace(token.TokenType); tokenType != "" {
		account.TokenType = tokenType
	}
	account.ExpiresAt = token.ExpiresAt()

	if validation != nil {
		if userID := strings.TrimSpace(validation.UserID); userID != "" {
			account.TwitchUserID = userID
		}
		if login := strings.TrimSpace(validation.Login); login != "" {
			account.Login = login
		}
		if len(validation.Scopes) > 0 {
			account.Scopes = append([]string(nil), validation.Scopes...)
		}
		account.LastValidatedAt = time.Now().UTC()
	}

	if err := accountStore.Save(ctx, *account); err != nil {
		return nil, err
	}

	return account, nil
}

func modeTokenNeedsRefresh(accessToken string, expiresAt time.Time) bool {
	if strings.TrimSpace(accessToken) == "" {
		return true
	}
	if expiresAt.IsZero() {
		return false
	}

	return time.Until(expiresAt) <= 5*time.Minute
}

func modeMissingScopes(actual []string, required ...string) []string {
	actualSet := make(map[string]struct{}, len(actual))
	for _, scope := range actual {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			continue
		}
		actualSet[scope] = struct{}{}
	}

	var missing []string
	for _, scope := range required {
		if _, ok := actualSet[scope]; !ok {
			missing = append(missing, scope)
		}
	}

	return missing
}
