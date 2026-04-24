package keywords

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/modules"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
)

const reloadInterval = 10 * time.Second

type Module struct {
	store      *postgres.KeywordStore
	defaults   *postgres.DefaultKeywordSettingStore
	game       *postgres.GameModuleSettingsStore
	gameName   func(context.Context) (string, error)
	nowPlaying *postgres.NowPlayingModuleSettingsStore
	songReply  func(context.Context) (string, error)
	state      *postgres.BotStateStore
	modes      *postgres.BotModeStore
	accounts   *postgres.TwitchAccountStore
	validator  SemanticValidator

	mu              sync.RWMutex
	keywords        []postgres.Keyword
	defaultSettings map[string]postgres.DefaultKeywordSetting
}

type SemanticValidator interface {
	ShouldTriggerKeyword(ctx context.Context, trigger, response, message string) (bool, float64, error)
}

type builtInKeyword struct {
	name      string
	candidate func(string) bool
	response  func(*Module, modules.CommandContext) (string, error)
}

var builtInKeywords = []builtInKeyword{
	{
		name:      "1v1",
		candidate: shouldHandleBuiltIn1v1,
		response: func(m *Module, ctx modules.CommandContext) (string, error) {
			return m.builtIn1v1Reply(ctx)
		},
	},
	{
		name:      "how do i join",
		candidate: shouldHandleBuiltInJoin,
		response: func(m *Module, ctx modules.CommandContext) (string, error) {
			return m.builtInJoinReply(ctx)
		},
	},
	{
		name:      "what song is this",
		candidate: shouldHandleBuiltInSong,
		response: func(m *Module, ctx modules.CommandContext) (string, error) {
			return m.builtInSongReply(ctx)
		},
	},
	{
		name:      "what game is this",
		candidate: shouldHandleBuiltInGame,
		response: func(m *Module, ctx modules.CommandContext) (string, error) {
			return m.builtInGameReply(ctx)
		},
	},
}

func New(store *postgres.KeywordStore, defaults *postgres.DefaultKeywordSettingStore, state *postgres.BotStateStore, accounts *postgres.TwitchAccountStore, allowedIDs ...string) *Module {
	_ = allowedIDs
	return &Module{
		store:           store,
		defaults:        defaults,
		state:           state,
		accounts:        accounts,
		defaultSettings: make(map[string]postgres.DefaultKeywordSetting),
	}
}

func DefaultSettingsDefaults() []postgres.DefaultKeywordSetting {
	return []postgres.DefaultKeywordSetting{
		{
			KeywordName:        "1v1",
			Enabled:            true,
			AIDetectionEnabled: true,
		},
		{
			KeywordName:        "how do i join",
			Enabled:            true,
			AIDetectionEnabled: true,
		},
		{
			KeywordName:        "what song is this",
			Enabled:            true,
			AIDetectionEnabled: true,
		},
		{
			KeywordName:        "what game is this",
			Enabled:            true,
			AIDetectionEnabled: true,
		},
	}
}

func (m *Module) SetModeStore(store *postgres.BotModeStore) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.modes = store
}

func (m *Module) SetGameModuleSettingsStore(store *postgres.GameModuleSettingsStore) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.game = store
}

func (m *Module) SetGameNameResolver(resolver func(context.Context) (string, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.gameName = resolver
}

func (m *Module) SetNowPlayingModuleSettingsStore(store *postgres.NowPlayingModuleSettingsStore) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.nowPlaying = store
}

func (m *Module) SetSongReplyResolver(resolver func(context.Context) (string, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.songReply = resolver
}

func (m *Module) Name() string {
	return "keywords"
}

func (m *Module) SetSemanticValidator(validator SemanticValidator) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.validator = validator
}

func (m *Module) RegisterCommands() map[string]modules.CommandDefinition {
	return map[string]modules.CommandDefinition{}
}

func (m *Module) Start(ctx context.Context) error {
	if m.defaults != nil {
		if err := m.defaults.EnsureDefaults(ctx, DefaultSettingsDefaults()); err != nil {
			return err
		}
	}
	if m.game != nil {
		if err := m.game.EnsureDefault(ctx); err != nil {
			return err
		}
	}
	if m.nowPlaying != nil {
		if err := m.nowPlaying.EnsureDefault(ctx); err != nil {
			return err
		}
	}
	if err := m.reload(ctx); err != nil {
		return err
	}

	go m.runReloadLoop(ctx)
	return nil
}

func (m *Module) HandleMessage(ctx modules.CommandContext) (modules.MessageResult, error) {
	if hasCommandPrefix(ctx.Message, ctx.CommandPrefix) {
		return modules.MessageResult{}, nil
	}

	message := strings.ToLower(strings.TrimSpace(ctx.Message))
	if message == "" {
		return modules.MessageResult{}, nil
	}

	if reply, ok, err := m.handleBuiltInKeywords(ctx, message); err != nil {
		return modules.MessageResult{}, err
	} else if ok {
		return modules.MessageResult{
			Reply:          reply,
			StopProcessing: true,
		}, nil
	}

	m.mu.RLock()
	keywords := append([]postgres.Keyword(nil), m.keywords...)
	validator := m.validator
	m.mu.RUnlock()

	candidates := make([]postgres.Keyword, 0, len(keywords))
	for _, keyword := range keywords {
		trigger := strings.TrimSpace(keyword.Trigger)
		if trigger == "" {
			continue
		}
		if keywordMatches(message, trigger) {
			candidates = append(candidates, keyword)
		}
	}

	if len(candidates) == 0 {
		return modules.MessageResult{}, nil
	}

	if validator == nil {
		return modules.MessageResult{
			Reply:          strings.TrimSpace(candidates[0].Response),
			StopProcessing: true,
		}, nil
	}

	bestReply := ""
	bestConfidence := -1.0
	for i, keyword := range candidates {
		if i >= 3 {
			break
		}

		shouldTrigger, confidence, err := validator.ShouldTriggerKeyword(context.Background(), keyword.Trigger, keyword.Response, ctx.Message)
		if err != nil {
			return modules.MessageResult{
				Reply:          strings.TrimSpace(candidates[0].Response),
				StopProcessing: true,
			}, nil
		}
		if !shouldTrigger {
			continue
		}
		if confidence > bestConfidence {
			bestConfidence = confidence
			bestReply = strings.TrimSpace(keyword.Response)
		}
	}

	if bestReply != "" {
		return modules.MessageResult{
			Reply:          bestReply,
			StopProcessing: true,
		}, nil
	}

	return modules.MessageResult{}, nil
}

func hasCommandPrefix(message, prefix string) bool {
	message = strings.TrimSpace(message)
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		prefix = "!"
	}

	return strings.HasPrefix(message, prefix)
}

func (m *Module) runReloadLoop(ctx context.Context) {
	ticker := time.NewTicker(reloadInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := m.reload(ctx); err != nil {
				fmt.Printf("keyword reload error: %v\n", err)
			}
		}
	}
}

func (m *Module) reload(ctx context.Context) error {
	if m.store == nil {
		return fmt.Errorf("keyword store is not configured")
	}

	items, err := m.store.List(ctx)
	if err != nil {
		return err
	}

	nextSettings := make(map[string]postgres.DefaultKeywordSetting)
	if m.defaults != nil {
		settings, err := m.defaults.List(ctx)
		if err != nil {
			return err
		}
		for _, setting := range settings {
			nextSettings[strings.ToLower(strings.TrimSpace(setting.KeywordName))] = setting
		}
	}

	m.mu.Lock()
	m.keywords = append([]postgres.Keyword(nil), items...)
	m.defaultSettings = nextSettings
	m.mu.Unlock()

	return nil
}

func (m *Module) handleBuiltInKeywords(ctx modules.CommandContext, message string) (string, bool, error) {
	for _, keyword := range builtInKeywords {
		setting := m.defaultKeywordSetting(keyword.name)
		if !setting.Enabled || !keyword.candidate(message) {
			continue
		}

		reply, err := keyword.response(m, ctx)
		if err != nil {
			return "", false, err
		}
		if strings.TrimSpace(reply) == "" {
			continue
		}

		if setting.AIDetectionEnabled && !m.shouldTriggerBuiltInKeyword(keyword.name, reply, ctx.Message) {
			continue
		}

		return reply, true, nil
	}

	return "", false, nil
}

func (m *Module) shouldTriggerBuiltInKeyword(keywordName, reply, message string) bool {
	m.mu.RLock()
	validator := m.validator
	m.mu.RUnlock()

	if validator == nil {
		return true
	}

	shouldTrigger, _, err := validator.ShouldTriggerKeyword(context.Background(), keywordName, reply, message)
	if err != nil {
		return true
	}

	return shouldTrigger
}

func (m *Module) defaultKeywordSetting(keywordName string) postgres.DefaultKeywordSetting {
	keywordName = strings.ToLower(strings.TrimSpace(keywordName))

	m.mu.RLock()
	setting, ok := m.defaultSettings[keywordName]
	m.mu.RUnlock()
	if ok {
		return setting
	}

	for _, item := range DefaultSettingsDefaults() {
		if strings.EqualFold(item.KeywordName, keywordName) {
			return item
		}
	}

	return postgres.DefaultKeywordSetting{
		KeywordName:        keywordName,
		Enabled:            true,
		AIDetectionEnabled: true,
	}
}

var simpleWordPattern = regexp.MustCompile(`^[a-z0-9_]+$`)
var oneVOneAskPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\bcan (i|we)\s+1v1(?:\b|s\b)`),
	regexp.MustCompile(`\bcould (i|we)\s+1v1(?:\b|s\b)`),
	regexp.MustCompile(`\bmay (i|we)\s+1v1(?:\b|s\b)`),
	regexp.MustCompile(`\bdo (you|u)\s+want to\s+1v1(?:\b|s\b)`),
	regexp.MustCompile(`\bwant to\s+1v1(?:\b|s\b)`),
	regexp.MustCompile(`\bwanna\s+1v1(?:\b|s\b)`),
	regexp.MustCompile(`\bare (you|u)\s+doing\s+1v1s?\b`),
	regexp.MustCompile(`\bdoing\s+1v1s?\b`),
	regexp.MustCompile(`\bwhen (are|is)\s+.*1v1`),
	regexp.MustCompile(`\b1v1s?\s*\?$`),
}

var joinAskPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\bhow (do|can) (i|we)\s+(join|get in)\b`),
	regexp.MustCompile(`\bcan (i|we)\s+join\b`),
	regexp.MustCompile(`\bhow do (i|we)\s+get in\b`),
	regexp.MustCompile(`\bhow to join\b`),
	regexp.MustCompile(`\b(join|server)\s+link\b`),
	regexp.MustCompile(`\bprivate server\b`),
	regexp.MustCompile(`\bhow can i play\b`),
}

var songAskPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\bwhat song is (this|that)\b`),
	regexp.MustCompile(`\bwhat music is (this|that)\b`),
	regexp.MustCompile(`\bwhat is this song\b`),
	regexp.MustCompile(`\bwhat'?s the song\b`),
	regexp.MustCompile(`\bname of (this|that) song\b`),
	regexp.MustCompile(`\bsong name\b`),
	regexp.MustCompile(`\bwhat track is (this|that)\b`),
}

var gameAskPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\bwhat game is (this|that)\b`),
	regexp.MustCompile(`\bwhat game are (you|u) playing\b`),
	regexp.MustCompile(`\bwhat are (you|u) playing\b`),
	regexp.MustCompile(`\bwhat game is (he|she|the streamer) playing\b`),
	regexp.MustCompile(`\bwhat'?s the game\b`),
	regexp.MustCompile(`\bwhich game is (this|that)\b`),
}

func (m *Module) builtIn1v1Reply(ctx modules.CommandContext) (string, error) {
	if m.state == nil {
		return "", nil
	}

	state, err := m.state.Get(context.Background())
	if err != nil {
		return "", err
	}

	target := targetMention(ctx)
	streamer := m.streamerName(context.Background())
	if state != nil && strings.EqualFold(strings.TrimSpace(state.CurrentModeKey), "1v1") {
		if reply, err := m.modeOwnedReply(context.Background(), state.CurrentModeKey, target, streamer, state.CurrentModeParam); err != nil {
			return "", err
		} else if strings.TrimSpace(reply) != "" {
			return reply, nil
		}

		return fmt.Sprintf("%s, type 1v1 in the chat ONCE to have a chance to 1v1 %s.", target, streamer), nil
	}

	return fmt.Sprintf("%s, %s is not currently doing 1v1s.", target, streamer), nil
}

func (m *Module) builtInJoinReply(ctx modules.CommandContext) (string, error) {
	if m.state == nil {
		return "", nil
	}

	state, err := m.state.Get(context.Background())
	if err != nil {
		return "", err
	}

	target := targetMention(ctx)
	streamer := m.streamerName(context.Background())
	modeKey := ""
	modeParam := ""
	if state != nil {
		modeKey = strings.ToLower(strings.TrimSpace(state.CurrentModeKey))
		modeParam = strings.TrimSpace(state.CurrentModeParam)
	}

	if reply, err := m.modeOwnedReply(context.Background(), modeKey, target, streamer, modeParam); err != nil {
		return "", err
	} else if strings.TrimSpace(reply) != "" {
		return reply, nil
	}

	switch modeKey {
	case "link":
		return fmt.Sprintf("%s, %s is currently using link mode. Use the website join panel or the posted private server link to get in.", target, streamer), nil
	case "join":
		return fmt.Sprintf("%s, %s is currently taking join requests. Use the active join keyword shown in chat once to get in.", target, streamer), nil
	case "1v1":
		return fmt.Sprintf("%s, %s is currently doing 1v1s. Type 1v1 in the chat ONCE to have a chance to 1v1 %s.", target, streamer, streamer), nil
	default:
		return fmt.Sprintf("%s, %s is not taking join requests right now.", target, streamer), nil
	}
}

func (m *Module) builtInGameReply(ctx modules.CommandContext) (string, error) {
	m.mu.RLock()
	gameStore := m.game
	gameNameResolver := m.gameName
	m.mu.RUnlock()

	responseTemplate := postgres.DefaultGameModuleSettings().KeywordResponse
	if gameStore != nil {
		settings, err := gameStore.Get(context.Background())
		if err != nil {
			return "", err
		}
		if settings != nil && strings.TrimSpace(settings.KeywordResponse) != "" {
			responseTemplate = settings.KeywordResponse
		}
	}

	gameName := "current game"
	if gameNameResolver != nil {
		resolvedName, err := gameNameResolver(context.Background())
		if err == nil && strings.TrimSpace(resolvedName) != "" {
			gameName = strings.TrimSpace(resolvedName)
		}
	}

	return m.renderModuleKeywordResponse(ctx, responseTemplate, map[string]string{
		"game": gameName,
	}), nil
}

func (m *Module) builtInSongReply(ctx modules.CommandContext) (string, error) {
	m.mu.RLock()
	nowPlayingStore := m.nowPlaying
	songReplyResolver := m.songReply
	m.mu.RUnlock()

	if songReplyResolver != nil {
		reply, err := songReplyResolver(context.Background())
		if err == nil && strings.TrimSpace(reply) != "" {
			return strings.TrimSpace(reply), nil
		}
	}

	responseTemplate := postgres.DefaultNowPlayingModuleSettings().KeywordResponse
	if nowPlayingStore != nil {
		settings, err := nowPlayingStore.Get(context.Background())
		if err != nil {
			return "", err
		}
		if settings != nil && strings.TrimSpace(settings.KeywordResponse) != "" {
			responseTemplate = settings.KeywordResponse
		}
	}

	return m.renderModuleKeywordResponse(ctx, responseTemplate, nil), nil
}

func (m *Module) renderModuleKeywordResponse(ctx modules.CommandContext, template string, extra map[string]string) string {
	target := targetMention(ctx)
	streamer := m.streamerName(context.Background())
	targetName := strings.TrimPrefix(strings.TrimSpace(target), "@")

	replacements := []string{
		"@{target}", "@" + targetName,
		"{target}", targetName,
		"{streamer}", strings.TrimSpace(streamer),
	}
	for key, value := range extra {
		normalized := strings.TrimSpace(key)
		if normalized == "" {
			continue
		}
		replacements = append(replacements, "{"+normalized+"}", strings.TrimSpace(value))
	}

	replacer := strings.NewReplacer(replacements...)

	return strings.TrimSpace(replacer.Replace(template))
}

func shouldHandleBuiltIn1v1(message string) bool {
	message = strings.ToLower(strings.TrimSpace(message))
	if message == "" || !strings.Contains(message, "1v1") {
		return false
	}

	switch message {
	case "1v1", "1v1 me", "1v1 pls", "1v1 please":
		return false
	}

	for _, pattern := range oneVOneAskPatterns {
		if pattern.MatchString(message) {
			return true
		}
	}

	return false
}

func shouldHandleBuiltInJoin(message string) bool {
	message = strings.ToLower(strings.TrimSpace(message))
	if message == "" {
		return false
	}
	if !strings.Contains(message, "join") && !strings.Contains(message, "server") && !strings.Contains(message, "link") {
		return false
	}

	for _, pattern := range joinAskPatterns {
		if pattern.MatchString(message) {
			return true
		}
	}

	return false
}

func shouldHandleBuiltInSong(message string) bool {
	message = strings.ToLower(strings.TrimSpace(message))
	if message == "" {
		return false
	}
	if !strings.Contains(message, "song") && !strings.Contains(message, "music") && !strings.Contains(message, "track") {
		return false
	}

	for _, pattern := range songAskPatterns {
		if pattern.MatchString(message) {
			return true
		}
	}

	return false
}

func shouldHandleBuiltInGame(message string) bool {
	message = strings.ToLower(strings.TrimSpace(message))
	if message == "" {
		return false
	}
	if !strings.Contains(message, "game") && !strings.Contains(message, "playing") {
		return false
	}

	for _, pattern := range gameAskPatterns {
		if pattern.MatchString(message) {
			return true
		}
	}

	return false
}

func targetMention(ctx modules.CommandContext) string {
	if name := strings.TrimSpace(ctx.Sender); name != "" {
		return "@" + name
	}
	if name := strings.TrimSpace(ctx.DisplayName); name != "" {
		return "@" + name
	}

	return "@there"
}

func (m *Module) streamerName(ctx context.Context) string {
	if m.accounts == nil {
		return "the streamer"
	}

	account, err := m.accounts.Get(ctx, postgres.TwitchAccountKindStreamer)
	if err != nil || account == nil {
		return "the streamer"
	}

	if strings.TrimSpace(account.DisplayName) != "" {
		return account.DisplayName
	}
	if strings.TrimSpace(account.Login) != "" {
		return account.Login
	}

	return "the streamer"
}

func (m *Module) modeOwnedReply(ctx context.Context, modeKey, target, streamer, modeParam string) (string, error) {
	modeKey = strings.TrimSpace(strings.ToLower(modeKey))
	if modeKey == "" {
		return "", nil
	}

	m.mu.RLock()
	modeStore := m.modes
	m.mu.RUnlock()
	if modeStore == nil {
		return "", nil
	}

	mode, err := modeStore.Get(ctx, modeKey)
	if err != nil || mode == nil {
		return "", err
	}

	template := strings.TrimSpace(mode.KeywordResponse)
	if template == "" {
		return "", nil
	}

	targetName := strings.TrimPrefix(strings.TrimSpace(target), "@")
	replacer := strings.NewReplacer(
		"@{target}", "@"+targetName,
		"{target}", targetName,
		"{streamer}", strings.TrimSpace(streamer),
		"{mode_param}", strings.TrimSpace(modeParam),
	)

	return strings.TrimSpace(replacer.Replace(template)), nil
}

func keywordMatches(message, trigger string) bool {
	message = strings.ToLower(strings.TrimSpace(message))
	trigger = strings.ToLower(strings.TrimSpace(trigger))
	if message == "" || trigger == "" {
		return false
	}

	if !simpleWordPattern.MatchString(trigger) {
		return strings.Contains(message, trigger)
	}

	start := 0
	for {
		index := strings.Index(message[start:], trigger)
		if index < 0 {
			return false
		}
		index += start

		beforeOK := index == 0 || !isWordChar(message[index-1])
		afterIndex := index + len(trigger)
		afterOK := afterIndex == len(message) || !isWordChar(message[afterIndex])
		if beforeOK && afterOK {
			return true
		}

		start = index + len(trigger)
		if start >= len(message) {
			return false
		}
	}
}

func isWordChar(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '_'
}
