package game

import (
	"context"
	"fmt"
	"regexp"
	"sort"
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
	twitchOAuth    *twitchoauth.Service
	twitchAccounts *postgres.TwitchAccountStore

	mu                 sync.RWMutex
	current            trackerState
	lastChannel        *twitchhelix.Channel
	lastChannelChecked time.Time
}

func New(cookie, twitchClientID, streamerID string, playtimeStore *postgres.RobloxPlaytimeStore, twitchOAuth *twitchoauth.Service, twitchAccounts *postgres.TwitchAccountStore) *Module {
	return &Module{
		configCookie:   strings.TrimSpace(cookie),
		twitchClientID: strings.TrimSpace(twitchClientID),
		streamerID:     strings.TrimSpace(streamerID),
		playtimeStore:  playtimeStore,
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
	if strings.TrimSpace(m.configCookie) == "" || strings.TrimSpace(m.twitchClientID) == "" || strings.TrimSpace(m.streamerID) == "" || m.playtimeStore == nil || m.twitchOAuth == nil {
		return nil
	}

	go m.runTracker(ctx)
	return nil
}

func (m *Module) playtime(ctx modules.CommandContext) (string, error) {
	_ = ctx

	current := m.currentState()
	if current.UniverseID == 0 {
		return "no Roblox experience is currently being tracked", nil
	}

	item, err := m.playtimeStore.Get(context.Background(), current.UniverseID)
	if err != nil {
		return "", err
	}

	total := time.Duration(0)
	if item != nil {
		total += time.Duration(item.TotalSeconds) * time.Second
	}
	if !current.TrackedAt.IsZero() {
		total += time.Since(current.TrackedAt)
	}

	gameName := cleanRobloxExperienceName(current.GameName)
	if gameName == "" && item != nil {
		gameName = cleanRobloxExperienceName(item.GameName)
	}
	if gameName == "" {
		gameName = "current Roblox experience"
	}

	return fmt.Sprintf("%s has been playing %s for %s.", m.streamerName(context.Background()), gameName, total.Round(time.Second)), nil
}

func (m *Module) game(ctx modules.CommandContext) (string, error) {
	_ = ctx

	channel, err := m.currentChannel(context.Background())
	if err != nil {
		return "", err
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
		return "", err
	}
	if experienceName == "" {
		return fmt.Sprintf("%s is currently playing a Roblox experience.", m.streamerName(context.Background())), nil
	}

	return fmt.Sprintf("%s is currently playing %s.", m.streamerName(context.Background()), experienceName), nil
}

func (m *Module) gamesPlayed(ctx modules.CommandContext) (string, error) {
	scope := "stream"
	if len(ctx.Args) > 0 {
		scope = strings.ToLower(strings.TrimSpace(ctx.Args[0]))
	}

	items, label, err := m.gamesPlayedScope(context.Background(), scope)
	if err != nil {
		return "", err
	}
	if len(items) == 0 {
		return "no Roblox games have been tracked yet", nil
	}

	parts := make([]string, 0, len(items))
	for _, item := range items {
		name := cleanRobloxExperienceName(item.GameName)
		if name == "" {
			name = fmt.Sprintf("Universe %d", item.UniverseID)
		}
		parts = append(parts, fmt.Sprintf("%s (%s)", name, (time.Duration(item.TotalSeconds)*time.Second).Round(time.Second)))
	}

	return label + ": " + strings.Join(parts, ", "), nil
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
				fmt.Printf("roblox tracker error: %v\n", err)
			}
		}
	}
}

func (m *Module) trackOnce(ctx context.Context) error {
	isRoblox, err := m.streamIsRoblox(ctx)
	if err != nil {
		return err
	}
	if !isRoblox {
		if err := m.flushCurrent(ctx, time.Now().UTC()); err != nil {
			return err
		}
		return nil
	}

	robloxClient := robloxapi.NewClient(nil, m.configCookie)
	authUser, err := robloxClient.GetAuthenticatedUser(ctx)
	if err != nil {
		return err
	}

	presences, err := robloxClient.GetPresences(ctx, []int64{authUser.ID})
	if err != nil {
		return err
	}
	if len(presences) == 0 {
		if err := m.flushCurrent(ctx, time.Now().UTC()); err != nil {
			return err
		}
		return nil
	}

	presence := presences[0]
	if presence.UniverseID == 0 {
		if err := m.flushCurrent(ctx, time.Now().UTC()); err != nil {
			return err
		}
		return nil
	}

	now := time.Now().UTC()
	next := trackerState{
		UniverseID:  presence.UniverseID,
		RootPlaceID: presence.RootPlaceID,
		GameName:    cleanRobloxExperienceName(presence.LastLocation),
		TrackedAt:   now,
	}

	prev := m.currentState()
	if prev.UniverseID != 0 && prev.UniverseID == next.UniverseID {
		next.StreamSessionID = prev.StreamSessionID
		if err := m.playtimeStore.AddDuration(ctx, prev.StreamSessionID, prev.UniverseID, prev.RootPlaceID, coalesceGameName(prev.GameName, next.GameName), prev.TrackedAt, now); err != nil {
			return err
		}
		m.setCurrent(next)
		return nil
	}

	if prev.UniverseID != 0 && !prev.TrackedAt.IsZero() {
		if err := m.playtimeStore.AddDuration(ctx, prev.StreamSessionID, prev.UniverseID, prev.RootPlaceID, prev.GameName, prev.TrackedAt, now); err != nil {
			return err
		}
	}

	if prev.StreamSessionID != "" {
		next.StreamSessionID = prev.StreamSessionID
	} else {
		next.StreamSessionID = newStreamSessionID(now)
	}

	m.setCurrent(next)
	return nil
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

func (m *Module) currentRobloxExperienceName(ctx context.Context) (string, error) {
	current := m.currentState()
	if strings.TrimSpace(current.GameName) != "" {
		return cleanRobloxExperienceName(current.GameName), nil
	}

	if strings.TrimSpace(m.configCookie) == "" {
		return "", nil
	}

	robloxClient := robloxapi.NewClient(nil, m.configCookie)
	authUser, err := robloxClient.GetAuthenticatedUser(ctx)
	if err != nil {
		return "", err
	}

	presences, err := robloxClient.GetPresences(ctx, []int64{authUser.ID})
	if err != nil {
		return "", err
	}
	if len(presences) == 0 {
		return "", nil
	}

	return cleanRobloxExperienceName(presences[0].LastLocation), nil
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

func coalesceGameName(values ...string) string {
	for _, value := range values {
		value = cleanRobloxExperienceName(value)
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

func (m *Module) gamesPlayedScope(ctx context.Context, scope string) ([]postgres.RobloxGamePlaytime, string, error) {
	now := time.Now().UTC()

	switch scope {
	case "", "stream":
		current := m.currentState()
		if current.StreamSessionID == "" {
			return nil, "Top 5 games played this stream", nil
		}

		items, err := m.playtimeStore.ListTopByStreamSession(ctx, current.StreamSessionID, 5)
		if err != nil {
			return nil, "", err
		}
		items = applyCurrentPlaytime(items, current, now)
		return items, "Top 5 games played this stream", nil
	case "all", "alltime":
		items, err := m.playtimeStore.ListTop(ctx, 5)
		if err != nil {
			return nil, "", err
		}
		items = applyCurrentPlaytime(items, m.currentState(), now)
		return items, "Top 5 games played all time", nil
	case "week":
		items, err := m.playtimeStore.ListTopByRange(ctx, now.AddDate(0, 0, -7), now, 5)
		return items, "Top 5 games played in the past week", err
	case "lastweek":
		start := now.AddDate(0, 0, -14)
		end := now.AddDate(0, 0, -7)
		items, err := m.playtimeStore.ListTopByRange(ctx, start, end, 5)
		return items, "Top 5 games played last week", err
	case "month":
		items, err := m.playtimeStore.ListTopByRange(ctx, now.AddDate(0, -1, 0), now, 5)
		return items, "Top 5 games played in the past month", err
	case "laststream":
		items, err := m.playtimeStore.ListTopByLastCompletedStream(ctx, 5)
		return items, "Top 5 games played last stream", err
	default:
		return nil, "", fmt.Errorf("unknown gamesplayed range %q", scope)
	}
}

func applyCurrentPlaytime(items []postgres.RobloxGamePlaytime, current trackerState, now time.Time) []postgres.RobloxGamePlaytime {
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
	if len(items) > 5 {
		items = items[:5]
	}

	return items
}

func newStreamSessionID(now time.Time) string {
	return fmt.Sprintf("roblox-stream-%d", now.UnixNano())
}
