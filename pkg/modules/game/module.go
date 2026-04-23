package game

import (
	"context"
	"fmt"
	"hash/fnv"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/modules"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	robloxapi "github.com/mr-cheeezz/dankbot/pkg/roblox/api"
	twitchhelix "github.com/mr-cheeezz/dankbot/pkg/twitch/helix"
	twitchoauth "github.com/mr-cheeezz/dankbot/pkg/twitch/oauth"
)

const robloxCategoryName = "roblox"

var (
	robloxNoiseTagPattern       = regexp.MustCompile(`(?i)\s*[\(\[]\s*(?:update|updated|new|alpha|beta|revamp|rework|fixed?|fixes?)\s*[\)\]]`)
	robloxLeadingNoiseWordRegex = regexp.MustCompile(`(?i)^(?:\s*(?:update|updated|new)\s*[-:|]?\s*)+`)
	robloxLeadingPromoRegex     = regexp.MustCompile(`(?i)^(?:\s*(?:[\[\(]?\s*)?(?:(?:spring|summer|autumn|fall|winter|halloween|christmas|xmas|holiday|easter|valentine'?s?|anniversary|major|mega)\s+)?(?:update|updated|new|revamp|rework|alpha|beta)\b(?:\s*[\]\)!:|!-]\s*)*)+`)
)

type trackerState struct {
	StreamSessionID string
	UniverseID      int64
	RootPlaceID     int64
	GameName        string
	TrackedAt       time.Time
}

type Module struct {
	configCookie   string
	twitchClientID string
	streamerID     string
	playtimeStore  *postgres.RobloxPlaytimeStore
	settingsStore  *postgres.GameModuleSettingsStore
	robloxAccounts *postgres.RobloxAccountStore
	twitchOAuth    *twitchoauth.Service
	twitchAccounts *postgres.TwitchAccountStore

	mu                 sync.RWMutex
	current            trackerState
	lastChannel        *twitchhelix.Channel
	lastChannelChecked time.Time
	lastLive           bool
	lastLiveChecked    time.Time
}

func New(
	cookie, twitchClientID, streamerID string,
	playtimeStore *postgres.RobloxPlaytimeStore,
	settingsStore *postgres.GameModuleSettingsStore,
	robloxAccounts *postgres.RobloxAccountStore,
	twitchOAuth *twitchoauth.Service,
	twitchAccounts *postgres.TwitchAccountStore,
) *Module {
	return &Module{
		configCookie:   strings.TrimSpace(cookie),
		twitchClientID: strings.TrimSpace(twitchClientID),
		streamerID:     strings.TrimSpace(streamerID),
		playtimeStore:  playtimeStore,
		settingsStore:  settingsStore,
		robloxAccounts: robloxAccounts,
		twitchOAuth:    twitchOAuth,
		twitchAccounts: twitchAccounts,
	}
}

func (m *Module) Name() string {
	return "game"
}

func (m *Module) RegisterCommands() map[string]modules.CommandDefinition {
	return map[string]modules.CommandDefinition{
		"game": {
			Handler:     m.game,
			Description: "Shows the current Twitch game, or the current Roblox experience when the stream category is Roblox.",
			Usage:       "!game",
			Example:     "!game",
		},
		"playtime": {
			Handler:     m.playtime,
			Description: "Shows how long the streamer has been playing the current tracked game.",
			Usage:       "!playtime",
			Example:     "!playtime",
		},
		"gamesplayed": {
			Handler:     m.gamesPlayed,
			Description: "Shows the top played games for the stream, week, month, or all time.",
			Usage:       "!gamesplayed [all|alltime|week|lastweek|month|laststream]",
			Example:     "!gamesplayed week",
		},
	}
}

func (m *Module) Start(ctx context.Context) error {
	if strings.TrimSpace(m.twitchClientID) == "" || strings.TrimSpace(m.streamerID) == "" || m.playtimeStore == nil || m.twitchOAuth == nil {
		return nil
	}

	go m.runTracker(ctx)
	return nil
}

func (m *Module) playtime(ctx modules.CommandContext) (string, error) {
	_ = ctx

	if live, err := m.streamIsLive(context.Background()); err == nil && !live {
		return fmt.Sprintf("%s is offline.", m.streamerName(context.Background())), nil
	}

	current := m.currentState()
	if current.UniverseID == 0 {
		return "No game is currently being tracked.", nil
	}

	total := time.Duration(0)
	var item *postgres.RobloxGamePlaytime
	if strings.TrimSpace(current.StreamSessionID) != "" {
		sessionItem, err := m.playtimeStore.GetByStreamSession(context.Background(), current.StreamSessionID, current.UniverseID)
		if err != nil {
			return "", err
		}
		item = sessionItem
		if item != nil {
			total += time.Duration(item.TotalSeconds) * time.Second
		}
	}
	if !current.TrackedAt.IsZero() {
		total += time.Since(current.TrackedAt)
	}

	gameName := displayTrackedGameName(current.UniverseID, current.GameName)
	if gameName == "" && item != nil {
		gameName = displayTrackedGameName(item.UniverseID, item.GameName)
	}
	if gameName == "" {
		gameName = "current stream game"
	}

	template := postgres.DefaultGameModuleSettings().PlaytimeTemplate
	if m.settingsStore != nil {
		if settings, err := m.settingsStore.Get(context.Background()); err != nil {
			return "", err
		} else if settings != nil && strings.TrimSpace(settings.PlaytimeTemplate) != "" {
			template = settings.PlaytimeTemplate
		}
	}

	replacer := strings.NewReplacer(
		"{streamer}", m.streamerName(context.Background()),
		"{game}", gameName,
		"{duration}", total.Round(time.Second).String(),
	)
	return replacer.Replace(template), nil
}

func (m *Module) game(ctx modules.CommandContext) (string, error) {
	_ = ctx

	channel, err := m.currentChannel(context.Background())
	if err != nil {
		return "current game unavailable", nil
	}
	if channel == nil {
		return "current game unavailable", nil
	}

	gameName := strings.TrimSpace(channel.GameName)
	if !strings.EqualFold(gameName, robloxCategoryName) {
		if gameName == "" {
			return "current game unavailable", nil
		}

		return fmt.Sprintf("%s is currently playing %s.", m.streamerName(context.Background()), gameName), nil
	}

	experienceName, err := m.currentRobloxExperienceName(context.Background())
	if err != nil {
		return fmt.Sprintf("%s is currently playing Roblox.", m.streamerName(context.Background())), nil
	}
	if experienceName == "" {
		return fmt.Sprintf("%s is currently playing a Roblox experience.", m.streamerName(context.Background())), nil
	}
	if strings.EqualFold(strings.TrimSpace(experienceName), "website") {
		return fmt.Sprintf("%s is switching games.", m.streamerName(context.Background())), nil
	}

	return fmt.Sprintf("%s is currently playing %s.", m.streamerName(context.Background()), experienceName), nil
}

func (m *Module) gamesPlayed(ctx modules.CommandContext) (string, error) {
	scope := "stream"
	if len(ctx.Args) > 0 {
		scope = strings.ToLower(strings.TrimSpace(ctx.Args[0]))
	}

	settings := postgres.DefaultGameModuleSettings()
	if m.settingsStore != nil {
		if stored, err := m.settingsStore.Get(context.Background()); err != nil {
			return "", err
		} else if stored != nil {
			settings = *stored
		}
	}

	limit := settings.GamesPlayedLimit
	items, label, err := m.gamesPlayedScope(context.Background(), scope, limit)
	if err != nil {
		return "", err
	}
	if len(items) == 0 {
		return "No games have been tracked yet.", nil
	}

	parts := make([]string, 0, len(items))
	for _, item := range items {
		name := displayTrackedGameName(item.UniverseID, item.GameName)
		if name == "" {
			name = "Unknown game"
		}
		duration := (time.Duration(item.TotalSeconds) * time.Second).Round(time.Second).String()
		itemReplacer := strings.NewReplacer(
			"{game}", name,
			"{duration}", duration,
		)
		parts = append(parts, itemReplacer.Replace(settings.GamesPlayedItemTemplate))
	}

	itemsText := strings.Join(parts, ", ")
	outReplacer := strings.NewReplacer(
		"{label}", label,
		"{items}", itemsText,
	)
	return outReplacer.Replace(settings.GamesPlayedTemplate), nil
}

func (m *Module) runTracker(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	_ = m.trackOnce(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := m.trackOnce(ctx); err != nil {
				fmt.Printf("game tracker error: %v\n", err)
			}
		}
	}
}

func (m *Module) trackOnce(ctx context.Context) error {
	channel, err := m.currentChannel(ctx)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	if channel == nil || strings.TrimSpace(channel.GameName) == "" {
		if err := m.flushCurrent(ctx, now); err != nil {
			return err
		}
		return nil
	}

	live, err := m.streamIsLive(ctx)
	if err == nil && !live {
		if err := m.flushCurrent(ctx, now); err != nil {
			return err
		}
		return nil
	}

	isRoblox := strings.EqualFold(strings.TrimSpace(channel.GameName), robloxCategoryName)
	if !isRoblox {
		next := trackerState{
			UniverseID: syntheticTwitchUniverseID(channel.GameID, channel.GameName),
			GameName:   strings.TrimSpace(channel.GameName),
			TrackedAt:  now,
		}
		m.advanceTracker(ctx, next, now)
		return nil
	}

	if strings.TrimSpace(m.configCookie) == "" {
		next := trackerState{
			UniverseID: syntheticTwitchUniverseID(channel.GameID, channel.GameName),
			GameName:   strings.TrimSpace(channel.GameName),
			TrackedAt:  now,
		}
		m.advanceTracker(ctx, next, now)
		return nil
	}

	robloxClient := robloxapi.NewClient(nil, m.configCookie)
	presenceUserID, err := m.presenceUserID(ctx, robloxClient)
	if err != nil {
		next := trackerState{
			UniverseID: syntheticTwitchUniverseID(channel.GameID, channel.GameName),
			GameName:   strings.TrimSpace(channel.GameName),
			TrackedAt:  now,
		}
		m.advanceTracker(ctx, next, now)
		return nil
	}

	presences, err := robloxClient.GetPresences(ctx, []int64{presenceUserID})
	if err != nil {
		next := trackerState{
			UniverseID: syntheticTwitchUniverseID(channel.GameID, channel.GameName),
			GameName:   strings.TrimSpace(channel.GameName),
			TrackedAt:  now,
		}
		m.advanceTracker(ctx, next, now)
		return nil
	}
	if len(presences) == 0 || presences[0].UniverseID == 0 {
		next := trackerState{
			UniverseID: syntheticTwitchUniverseID(channel.GameID, channel.GameName),
			GameName:   strings.TrimSpace(channel.GameName),
			TrackedAt:  now,
		}
		m.advanceTracker(ctx, next, now)
		return nil
	}

	presence := presences[0]
	next := trackerState{
		UniverseID:  presence.UniverseID,
		RootPlaceID: presence.RootPlaceID,
		GameName:    cleanRobloxExperienceName(presence.LastLocation),
		TrackedAt:   now,
	}
	m.advanceTracker(ctx, next, now)
	return nil
}

func (m *Module) advanceTracker(ctx context.Context, next trackerState, now time.Time) {
	if next.UniverseID == 0 {
		if err := m.flushCurrent(ctx, time.Now().UTC()); err != nil {
			fmt.Printf("game tracker flush error: %v\n", err)
		}
		return
	}

	prev := m.currentState()
	if prev.UniverseID != 0 && prev.UniverseID == next.UniverseID {
		next.StreamSessionID = prev.StreamSessionID
		if err := m.playtimeStore.AddDuration(ctx, prev.StreamSessionID, prev.UniverseID, prev.RootPlaceID, coalesceGameName(prev.UniverseID, prev.GameName, next.GameName), prev.TrackedAt, now); err != nil {
			fmt.Printf("game tracker add duration error: %v\n", err)
			return
		}
		m.setCurrent(next)
		return
	}

	if prev.UniverseID != 0 && !prev.TrackedAt.IsZero() {
		if err := m.playtimeStore.AddDuration(ctx, prev.StreamSessionID, prev.UniverseID, prev.RootPlaceID, prev.GameName, prev.TrackedAt, now); err != nil {
			fmt.Printf("game tracker rollover add duration error: %v\n", err)
		}
	}

	if prev.StreamSessionID != "" {
		next.StreamSessionID = prev.StreamSessionID
	} else {
		next.StreamSessionID = newStreamSessionID(now)
	}

	m.setCurrent(next)
}

func (m *Module) streamIsRoblox(ctx context.Context) (bool, error) {
	channel, err := m.currentChannel(ctx)
	if err != nil {
		return false, err
	}
	if channel == nil {
		return false, nil
	}

	return strings.EqualFold(strings.TrimSpace(channel.GameName), robloxCategoryName), nil
}

func (m *Module) currentChannel(ctx context.Context) (*twitchhelix.Channel, error) {
	m.mu.RLock()
	if m.lastChannel != nil && time.Since(m.lastChannelChecked) < 30*time.Second {
		channelCopy := *m.lastChannel
		m.mu.RUnlock()
		return &channelCopy, nil
	}
	m.mu.RUnlock()

	token, err := m.twitchOAuth.AppToken(ctx)
	if err != nil {
		return m.cachedChannelOnError(err)
	}

	helixClient := twitchhelix.NewClient(m.twitchClientID, token.AccessToken)
	channels, err := helixClient.GetChannelsByBroadcasterIDs(ctx, []string{m.streamerID})
	if err != nil {
		return m.cachedChannelOnError(err)
	}
	if len(channels) == 0 {
		m.mu.Lock()
		m.lastChannel = nil
		m.lastChannelChecked = time.Now().UTC()
		m.mu.Unlock()
		return nil, nil
	}

	m.mu.Lock()
	channelCopy := channels[0]
	m.lastChannel = &channelCopy
	m.lastChannelChecked = time.Now().UTC()
	m.mu.Unlock()

	return &channels[0], nil
}

func (m *Module) streamIsLive(ctx context.Context) (bool, error) {
	if strings.TrimSpace(m.twitchClientID) == "" || strings.TrimSpace(m.streamerID) == "" || m.twitchOAuth == nil {
		// If we can't check, don't block the command output.
		return true, nil
	}

	m.mu.RLock()
	if !m.lastLiveChecked.IsZero() && time.Since(m.lastLiveChecked) < 30*time.Second {
		cached := m.lastLive
		m.mu.RUnlock()
		return cached, nil
	}
	m.mu.RUnlock()

	token, err := m.twitchOAuth.AppToken(ctx)
	if err != nil {
		return true, err
	}

	helixClient := twitchhelix.NewClient(m.twitchClientID, token.AccessToken)
	streams, err := helixClient.GetStreamsByUserIDs(ctx, []string{m.streamerID})
	if err != nil {
		return true, err
	}

	live := len(streams) > 0
	m.mu.Lock()
	m.lastLive = live
	m.lastLiveChecked = time.Now().UTC()
	m.mu.Unlock()

	return live, nil
}

func (m *Module) currentRobloxExperienceName(ctx context.Context) (string, error) {
	current := m.currentState()
	if strings.TrimSpace(current.GameName) != "" {
		return cleanRobloxExperienceName(current.GameName), nil
	}

	if strings.TrimSpace(m.configCookie) == "" {
		return "", nil
	}

	robloxClient := robloxapi.NewClient(nil, m.configCookie)
	presenceUserID, err := m.presenceUserID(ctx, robloxClient)
	if err != nil {
		return "", err
	}

	presences, err := robloxClient.GetPresences(ctx, []int64{presenceUserID})
	if err != nil {
		return "", err
	}
	if len(presences) == 0 {
		return "", nil
	}

	return cleanRobloxExperienceName(presences[0].LastLocation), nil
}

func (m *Module) presenceUserID(ctx context.Context, robloxClient *robloxapi.Client) (int64, error) {
	if m.robloxAccounts != nil {
		account, err := m.robloxAccounts.Get(ctx, postgres.RobloxAccountKindStreamer)
		if err == nil && account != nil {
			if linkedID, parseErr := strconv.ParseInt(strings.TrimSpace(account.RobloxUserID), 10, 64); parseErr == nil && linkedID > 0 {
				return linkedID, nil
			}
		}
	}

	authUser, err := robloxClient.GetAuthenticatedUser(ctx)
	if err != nil {
		return 0, err
	}
	if authUser == nil || authUser.ID <= 0 {
		return 0, fmt.Errorf("roblox authenticated user id is missing")
	}

	return authUser.ID, nil
}

func (m *Module) currentState() trackerState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.current
}

func (m *Module) setCurrent(state trackerState) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.current = state
}

func (m *Module) clearCurrent() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.current = trackerState{}
}

func (m *Module) cachedChannelOnError(err error) (*twitchhelix.Channel, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.lastChannel != nil {
		channelCopy := *m.lastChannel
		return &channelCopy, nil
	}

	return nil, err
}

func (m *Module) flushCurrent(ctx context.Context, now time.Time) error {
	prev := m.currentState()
	if prev.UniverseID != 0 && !prev.TrackedAt.IsZero() {
		if err := m.playtimeStore.AddDuration(ctx, prev.StreamSessionID, prev.UniverseID, prev.RootPlaceID, prev.GameName, prev.TrackedAt, now); err != nil {
			return err
		}
	}

	m.clearCurrent()
	return nil
}

func coalesceGameName(universeID int64, values ...string) string {
	for _, value := range values {
		if isSyntheticTwitchUniverseID(universeID) {
			value = strings.TrimSpace(value)
		} else {
			value = cleanRobloxExperienceName(value)
		}
		if value != "" {
			return value
		}
	}

	return ""
}

func (m *Module) streamerName(ctx context.Context) string {
	if m.twitchAccounts == nil {
		return "the streamer"
	}

	account, err := m.twitchAccounts.Get(ctx, postgres.TwitchAccountKindStreamer)
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

func cleanRobloxExperienceName(name string) string {
	original := strings.TrimSpace(name)
	if original == "" {
		return ""
	}

	cleaned := robloxNoiseTagPattern.ReplaceAllString(original, " ")
	cleaned = robloxLeadingPromoRegex.ReplaceAllString(cleaned, "")
	cleaned = robloxLeadingNoiseWordRegex.ReplaceAllString(cleaned, "")
	cleaned = strings.Join(strings.Fields(cleaned), " ")
	cleaned = strings.Trim(cleaned, " -:|[]()")
	cleaned = strings.Join(strings.Fields(cleaned), " ")
	if cleaned == "" {
		return original
	}

	return cleaned
}

func (m *Module) gamesPlayedScope(ctx context.Context, scope string, limit int) ([]postgres.RobloxGamePlaytime, string, error) {
	now := time.Now().UTC()
	if limit < 1 {
		limit = postgres.DefaultGameModuleSettings().GamesPlayedLimit
	}

	switch scope {
	case "", "stream":
		current := m.currentState()
		if current.StreamSessionID == "" {
			return nil, fmt.Sprintf("Top %d games played this stream", limit), nil
		}

		items, err := m.playtimeStore.ListTopByStreamSession(ctx, current.StreamSessionID, limit)
		if err != nil {
			return nil, "", err
		}
		items = applyCurrentPlaytime(items, current, now, limit)
		return items, fmt.Sprintf("Top %d games played this stream", limit), nil
	case "all", "alltime":
		items, err := m.playtimeStore.ListTop(ctx, limit)
		if err != nil {
			return nil, "", err
		}
		items = applyCurrentPlaytime(items, m.currentState(), now, limit)
		return items, fmt.Sprintf("Top %d games played all time", limit), nil
	case "week":
		items, err := m.playtimeStore.ListTopByRange(ctx, now.AddDate(0, 0, -7), now, limit)
		return items, fmt.Sprintf("Top %d games played in the past week", limit), err
	case "lastweek":
		start := now.AddDate(0, 0, -14)
		end := now.AddDate(0, 0, -7)
		items, err := m.playtimeStore.ListTopByRange(ctx, start, end, limit)
		return items, fmt.Sprintf("Top %d games played last week", limit), err
	case "month":
		items, err := m.playtimeStore.ListTopByRange(ctx, now.AddDate(0, -1, 0), now, limit)
		return items, fmt.Sprintf("Top %d games played in the past month", limit), err
	case "laststream":
		items, err := m.playtimeStore.ListTopByLastCompletedStream(ctx, limit)
		return items, fmt.Sprintf("Top %d games played last stream", limit), err
	default:
		return nil, "", fmt.Errorf("unknown gamesplayed range %q", scope)
	}
}

func applyCurrentPlaytime(items []postgres.RobloxGamePlaytime, current trackerState, now time.Time, limit int) []postgres.RobloxGamePlaytime {
	if current.UniverseID == 0 || current.TrackedAt.IsZero() {
		return items
	}

	extraSeconds := int64(now.Sub(current.TrackedAt) / time.Second)
	if extraSeconds <= 0 {
		return items
	}

	items = append([]postgres.RobloxGamePlaytime(nil), items...)
	for i := range items {
		if items[i].UniverseID == current.UniverseID {
			items[i].TotalSeconds += extraSeconds
			if strings.TrimSpace(items[i].GameName) == "" {
				items[i].GameName = current.GameName
			}
			return items
		}
	}

	items = append(items, postgres.RobloxGamePlaytime{
		UniverseID:   current.UniverseID,
		RootPlaceID:  current.RootPlaceID,
		GameName:     current.GameName,
		TotalSeconds: extraSeconds,
	})

	sort.Slice(items, func(i, j int) bool {
		if items[i].TotalSeconds == items[j].TotalSeconds {
			return items[i].UniverseID < items[j].UniverseID
		}
		return items[i].TotalSeconds > items[j].TotalSeconds
	})
	if limit < 1 {
		limit = postgres.DefaultGameModuleSettings().GamesPlayedLimit
	}
	if len(items) > limit {
		items = items[:limit]
	}

	return items
}

func newStreamSessionID(now time.Time) string {
	return fmt.Sprintf("game-stream-%d", now.UnixNano())
}

func syntheticTwitchUniverseID(gameID, gameName string) int64 {
	seed := strings.TrimSpace(gameID)
	if seed == "" {
		seed = strings.TrimSpace(strings.ToLower(gameName))
	}
	if seed == "" {
		seed = "twitch-unknown"
	}
	hasher := fnv.New64a()
	_, _ = hasher.Write([]byte(seed))
	id := int64(hasher.Sum64())
	if id > 0 {
		id = -id
	}
	if id == 0 {
		id = -1
	}
	return id
}

func isSyntheticTwitchUniverseID(universeID int64) bool {
	return universeID < 0
}

func displayTrackedGameName(universeID int64, name string) string {
	if isSyntheticTwitchUniverseID(universeID) {
		return strings.TrimSpace(name)
	}
	return cleanRobloxExperienceName(name)
}
