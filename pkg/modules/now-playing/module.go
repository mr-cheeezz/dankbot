package spotify

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/modules"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	spotifyapi "github.com/mr-cheeezz/dankbot/pkg/spotify/api"
	spotifyoauth "github.com/mr-cheeezz/dankbot/pkg/spotify/oauth"
)

type Module struct {
	accounts       *postgres.SpotifyAccountStore
	twitchAccounts *postgres.TwitchAccountStore
	settings       *postgres.NowPlayingModuleSettingsStore
	auditStore     *postgres.AuditLogStore
	oauth          *spotifyoauth.Service
	allowedIDs     map[string]struct{}
	isLive         func(context.Context) (bool, error)

	mu                 sync.RWMutex
	channel            string
	say                func(channel, message string) error
	lastAnnouncedTrack string
	initializedTrack   bool
}

var (
	spotifyFeaturingTagPattern = regexp.MustCompile(`(?i)\s*[\(\[][^)\]]*\b(?:feat\.?|ft\.?|featuring)\b[^)\]]*[\)\]]`)
	spotifyTrailingFeatPattern = regexp.MustCompile(`(?i)\s*[-–—]\s*(?:feat\.?|ft\.?|featuring)\b.*$`)
	spotifyNoiseTagPattern     = regexp.MustCompile(`(?i)\s*[\(\[][^)\]]*\b(?:official(?:\s+music)?\s+video|lyric\s+video|audio|visualizer|remaster(?:ed)?(?:\s+\d{4})?|deluxe|bonus\s+track|radio\s+edit|extended|mono|stereo|clean|explicit|album\s+version|single\s+version)\b[^)\]]*[\)\]]`)
)

func New(accounts *postgres.SpotifyAccountStore, twitchAccounts *postgres.TwitchAccountStore, auditStore *postgres.AuditLogStore, oauthService *spotifyoauth.Service, allowedIDs ...string) *Module {
	allowed := make(map[string]struct{})
	for _, id := range allowedIDs {
		id = strings.TrimSpace(id)
		if id != "" {
			allowed[id] = struct{}{}
		}
	}

	return &Module{
		accounts:       accounts,
		twitchAccounts: twitchAccounts,
		auditStore:     auditStore,
		oauth:          oauthService,
		allowedIDs:     allowed,
	}
}

func (m *Module) Name() string {
	return "now-playing"
}

func (m *Module) RegisterCommands() map[string]modules.CommandDefinition {
	return map[string]modules.CommandDefinition{
		"song": {
			Handler:     m.song,
			Description: "Shows Spotify playback info and lets mods control playback.",
			Usage:       "!song [next|last|add|skip|volume]",
			Example:     "!song next",
		},
	}
}

func (m *Module) Start(ctx context.Context) error {
	go m.runAnnouncements(ctx)
	return nil
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

func (m *Module) SetSettingsStore(store *postgres.NowPlayingModuleSettingsStore) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.settings = store
}

func (m *Module) SongKeywordReply(ctx context.Context) (string, error) {
	return m.currentSong(ctx)
}

func (m *Module) song(ctx modules.CommandContext) (string, error) {
	settings, err := m.getSettings(context.Background())
	if err != nil {
		return "", err
	}

	if len(ctx.Args) == 0 {
		if !settings.SongCommandEnabled {
			return "", nil
		}
		return m.currentSong(context.Background())
	}

	switch strings.ToLower(strings.TrimSpace(ctx.Args[0])) {
	case "next":
		if !settings.SongNextCommandEnabled {
			return "", nil
		}
		return m.nextSong(context.Background())
	case "last":
		if !settings.SongLastCommandEnabled {
			return "", nil
		}
		return m.lastSong(context.Background())
	case "add":
		if !m.canControlPlayback(ctx) {
			return "", nil
		}
		if len(ctx.Args) < 2 {
			return "usage: " + commandPrefix(ctx) + "song add <spotify url|spotify uri|search terms>", nil
		}
		return m.addSong(ctx, strings.Join(ctx.Args[1:], " "))
	case "skip":
		if !m.canControlPlayback(ctx) {
			return "", nil
		}
		return m.skipSong(ctx)
	case "volume":
		if !m.canControlPlayback(ctx) {
			return "", nil
		}
		if len(ctx.Args) < 2 {
			return "usage: " + commandPrefix(ctx) + "song volume <0-100>", nil
		}
		return m.setVolume(ctx, ctx.Args[1])
	default:
		return "usage: " + commandPrefix(ctx) + "song [next|last|add|skip|volume]", nil
	}
}

func (m *Module) currentSong(ctx context.Context) (string, error) {
	client, err := m.client(ctx)
	if err != nil {
		return "", err
	}

	playing, err := client.GetCurrentlyPlaying(ctx, "")
	if err != nil {
		return "", err
	}

	streamer := m.streamerName(ctx)
	if playing == nil || playing.Item == nil {
		return fmt.Sprintf("%s is not listening to anything right now.", streamer), nil
	}

	return fmt.Sprintf("%s is currently listening to %s.", streamer, formatTrack(*playing.Item)), nil
}

func (m *Module) nextSong(ctx context.Context) (string, error) {
	client, err := m.client(ctx)
	if err != nil {
		return "", err
	}

	queue, err := client.GetQueue(ctx)
	if err != nil {
		return "", err
	}

	if queue == nil || len(queue.Queue) == 0 {
		return "There isn't a next song in the queue right now.", nil
	}

	return fmt.Sprintf("The next song in the queue is %s.", formatTrack(queue.Queue[0])), nil
}

func (m *Module) lastSong(ctx context.Context) (string, error) {
	client, err := m.client(ctx)
	if err != nil {
		return "", err
	}

	page, err := client.GetRecentlyPlayed(ctx, 1, "", "")
	if err != nil {
		return "", err
	}

	streamer := m.streamerName(ctx)
	if page == nil || len(page.Items) == 0 {
		return fmt.Sprintf("I couldn't find the last song %s listened to.", streamer), nil
	}

	return fmt.Sprintf("The last song %s listened to was %s.", streamer, formatTrack(page.Items[0].Track)), nil
}

func (m *Module) addSong(ctx modules.CommandContext, input string) (string, error) {
	client, err := m.client(context.Background())
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return "usage: " + commandPrefix(ctx) + "song add <spotify url|spotify uri|search terms>", nil
	}

	trackURI, title, err := resolveTrackURI(context.Background(), client, input)
	if err != nil {
		return "", err
	}
	if title == "" {
		if resolvedTitle, resolveErr := resolveTrackTitleFromURI(context.Background(), client, trackURI); resolveErr == nil {
			title = resolvedTitle
		}
	}

	if err := client.AddToQueue(context.Background(), trackURI, ""); err != nil {
		return "", err
	}
	m.logAction(ctx, commandPrefix(ctx)+"song add", buildAddSongAuditDetail(title, input))

	if title != "" {
		return fmt.Sprintf("Successfully added %s to the queue.", queueTrackLabel(title)), nil
	}

	return "Successfully added the requested song to the queue.", nil
}

func (m *Module) skipSong(ctx modules.CommandContext) (string, error) {
	client, err := m.client(context.Background())
	if err != nil {
		return "", err
	}

	if err := client.SkipNext(context.Background(), ""); err != nil {
		return "", err
	}
	m.logAction(ctx, commandPrefix(ctx)+"song skip", "skipped the current spotify song")

	return "Skipped the current song.", nil
}

func (m *Module) setVolume(ctx modules.CommandContext, raw string) (string, error) {
	client, err := m.client(context.Background())
	if err != nil {
		return "", err
	}

	volume, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return "usage: " + commandPrefix(ctx) + "song volume <0-100>", nil
	}

	if err := client.SetVolume(context.Background(), volume, ""); err != nil {
		return "", err
	}
	m.logAction(ctx, commandPrefix(ctx)+"song volume", fmt.Sprintf("set spotify volume to %d%%", volume))

	return fmt.Sprintf("Set Spotify volume to %d%%.", volume), nil
}

func (m *Module) client(ctx context.Context) (*spotifyapi.Client, error) {
	account, err := m.linkedAccount(ctx)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, fmt.Errorf("spotify streamer account is not linked yet")
	}
	if strings.TrimSpace(account.AccessToken) == "" {
		return nil, fmt.Errorf("spotify streamer access token is missing")
	}

	return spotifyapi.NewClient(nil, account.AccessToken), nil
}

func (m *Module) linkedAccount(ctx context.Context) (*postgres.SpotifyAccount, error) {
	if m.accounts == nil {
		return nil, fmt.Errorf("spotify account store is not configured")
	}

	account, err := m.accounts.Get(ctx, postgres.SpotifyAccountKindStreamer)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, fmt.Errorf("spotify streamer account is not linked yet")
	}

	if !tokenNeedsRefresh(account.AccessToken, account.ExpiresAt) || strings.TrimSpace(account.RefreshToken) == "" || m.oauth == nil {
		if strings.TrimSpace(account.AccessToken) == "" {
			return nil, fmt.Errorf("spotify streamer access token is missing")
		}
		return account, nil
	}

	token, err := m.oauth.RefreshToken(ctx, account.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("refresh spotify token: %w", err)
	}

	account.AccessToken = strings.TrimSpace(token.AccessToken)
	if refreshToken := strings.TrimSpace(token.RefreshToken); refreshToken != "" {
		account.RefreshToken = refreshToken
	}
	if scope := strings.TrimSpace(token.Scope); scope != "" {
		account.Scope = scope
	}
	if tokenType := strings.TrimSpace(token.TokenType); tokenType != "" {
		account.TokenType = tokenType
	}
	account.ExpiresAt = token.ExpiresAt()

	if err := m.accounts.Save(ctx, *account); err != nil {
		return nil, err
	}

	return account, nil
}

func (m *Module) streamerName(ctx context.Context) string {
	if m.twitchAccounts == nil {
		return "Streamer"
	}

	account, err := m.twitchAccounts.Get(ctx, postgres.TwitchAccountKindStreamer)
	if err != nil || account == nil {
		return "Streamer"
	}

	if strings.TrimSpace(account.DisplayName) != "" {
		return account.DisplayName
	}
	if strings.TrimSpace(account.Login) != "" {
		return account.Login
	}

	return "Streamer"
}

func (m *Module) canControlPlayback(ctx modules.CommandContext) bool {
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

	_, ok := m.allowedIDs[senderID]
	return ok
}

func (m *Module) actorName(ctx modules.CommandContext) string {
	if name := strings.TrimSpace(ctx.DisplayName); name != "" {
		return name
	}
	if name := strings.TrimSpace(ctx.Sender); name != "" {
		return name
	}

	return strings.TrimSpace(ctx.SenderID)
}

func commandPrefix(ctx modules.CommandContext) string {
	prefix := strings.TrimSpace(ctx.CommandPrefix)
	if prefix == "" {
		return "!"
	}

	return prefix
}

func (m *Module) logAction(ctx modules.CommandContext, command, detail string) {
	if m.auditStore == nil || strings.TrimSpace(command) == "" || strings.TrimSpace(detail) == "" {
		return
	}

	if _, err := m.auditStore.Create(context.Background(), postgres.AuditLog{
		Platform:  strings.TrimSpace(ctx.Platform),
		ActorID:   strings.TrimSpace(ctx.SenderID),
		ActorName: m.actorName(ctx),
		Command:   strings.TrimSpace(command),
		Detail:    strings.TrimSpace(detail),
	}); err != nil {
		fmt.Printf("audit log error: %v\n", err)
	}
}

func buildAddSongAuditDetail(title, input string) string {
	title = strings.TrimSpace(title)
	input = strings.TrimSpace(input)

	if title != "" {
		return fmt.Sprintf("added %s to the spotify queue", title)
	}
	if input != "" {
		return fmt.Sprintf("queued a spotify request from %s", input)
	}

	return "added a song to the spotify queue"
}

func (m *Module) runAnnouncements(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	_ = m.tickAnnouncement(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := m.tickAnnouncement(ctx); err != nil {
				fmt.Printf("spotify announcer error: %v\n", err)
			}
		}
	}
}

func (m *Module) tickAnnouncement(ctx context.Context) error {
	channel, say, checker := m.output()
	if channel == "" || say == nil {
		return nil
	}
	if checker != nil {
		live, err := checker(ctx)
		if err != nil {
			return err
		}
		if !live {
			return nil
		}
	}

	client, err := m.client(ctx)
	if err != nil {
		return nil
	}

	playing, err := client.GetCurrentlyPlaying(ctx, "")
	if err != nil {
		return err
	}
	if playing == nil || !playing.IsPlaying || playing.Item == nil {
		return nil
	}

	trackID := strings.TrimSpace(playing.Item.ID)
	if trackID == "" {
		trackID = strings.TrimSpace(playing.Item.URI)
	}
	if trackID == "" {
		return nil
	}

	streamer := m.streamerName(ctx)
	track := formatTrack(*playing.Item)
	settings, err := m.getSettings(ctx)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initializedTrack {
		m.lastAnnouncedTrack = trackID
		m.initializedTrack = true
		return nil
	}
	if m.lastAnnouncedTrack == trackID {
		return nil
	}

	m.lastAnnouncedTrack = trackID
	return say(channel, renderSongChangeTemplate(settings.SongChangeMessageTemplate, streamer, track))
}

func (m *Module) output() (string, func(channel, message string) error, func(context.Context) (bool, error)) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.channel, m.say, m.isLive
}

func (m *Module) getSettings(ctx context.Context) (postgres.NowPlayingModuleSettings, error) {
	defaults := postgres.DefaultNowPlayingModuleSettings()

	m.mu.RLock()
	settingsStore := m.settings
	m.mu.RUnlock()
	if settingsStore == nil {
		return defaults, nil
	}

	settings, err := settingsStore.Get(ctx)
	if err != nil {
		return defaults, err
	}
	if settings == nil {
		return defaults, nil
	}

	return *settings, nil
}

func renderSongChangeTemplate(template, streamer, song string) string {
	template = strings.TrimSpace(template)
	if template == "" {
		template = postgres.DefaultNowPlayingModuleSettings().SongChangeMessageTemplate
	}

	replacer := strings.NewReplacer(
		"{streamer}", strings.TrimSpace(streamer),
		"{song}", strings.TrimSpace(song),
		"{track}", strings.TrimSpace(song),
	)
	return strings.TrimSpace(replacer.Replace(template))
}

func resolveTrackURI(ctx context.Context, client *spotifyapi.Client, input string) (string, string, error) {
	if uri, ok := spotifyTrackURI(input); ok {
		if trackTitle, err := resolveTrackTitleFromURI(ctx, client, uri); err == nil {
			return uri, trackTitle, nil
		}
		return uri, "", nil
	}

	tracks, err := client.SearchTracks(ctx, input, 1)
	if err != nil {
		return "", "", err
	}
	if len(tracks) == 0 {
		return "", "", fmt.Errorf("couldn't find that song on spotify")
	}

	return tracks[0].URI, formatTrack(tracks[0]), nil
}

func resolveTrackTitleFromURI(ctx context.Context, client *spotifyapi.Client, uri string) (string, error) {
	trackID, ok := spotifyTrackIDFromURI(uri)
	if !ok {
		return "", fmt.Errorf("spotify track id not found")
	}

	track, err := client.GetTrack(ctx, trackID)
	if err != nil {
		return "", err
	}
	if track == nil {
		return "", fmt.Errorf("spotify track lookup returned no data")
	}

	return formatTrack(*track), nil
}

func spotifyTrackURI(input string) (string, bool) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", false
	}

	lower := strings.ToLower(input)
	if strings.HasPrefix(lower, "spotify:track:") {
		return input, true
	}

	parsed, err := url.Parse(input)
	if err != nil {
		return "", false
	}

	host := strings.ToLower(parsed.Host)
	if !strings.Contains(host, "spotify.com") {
		return "", false
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	for index := 0; index < len(parts)-1; index++ {
		if parts[index] == "track" && strings.TrimSpace(parts[index+1]) != "" {
			return "spotify:track:" + strings.TrimSpace(parts[index+1]), true
		}
	}

	return "", false
}

func spotifyTrackIDFromURI(uri string) (string, bool) {
	uri = strings.TrimSpace(uri)
	if uri == "" {
		return "", false
	}

	lower := strings.ToLower(uri)
	if !strings.HasPrefix(lower, "spotify:track:") {
		return "", false
	}

	parts := strings.SplitN(uri, ":", 3)
	if len(parts) != 3 {
		return "", false
	}

	id := strings.TrimSpace(parts[2])
	if id == "" {
		return "", false
	}

	return id, true
}

func formatTrack(track spotifyapi.Track) string {
	artistNames := make([]string, 0, len(track.Artists))
	for _, artist := range track.Artists {
		if strings.TrimSpace(artist.Name) != "" {
			artistNames = append(artistNames, artist.Name)
		}
	}

	title := cleanSpotifyTrackTitle(track.Name)
	if title == "" {
		title = "unknown track"
	}

	if len(artistNames) == 0 {
		return title
	}

	return fmt.Sprintf("%s by %s", title, strings.Join(artistNames, ", "))
}

func queueTrackLabel(formattedTrack string) string {
	track := strings.TrimSpace(formattedTrack)
	if track == "" {
		return "the requested song"
	}

	parts := strings.SplitN(track, " by ", 2)
	if len(parts) != 2 {
		return track
	}

	song := strings.TrimSpace(parts[0])
	artist := strings.TrimSpace(parts[1])
	if song == "" || artist == "" {
		return track
	}

	return fmt.Sprintf("%s - %s", song, artist)
}

func cleanSpotifyTrackTitle(title string) string {
	original := strings.TrimSpace(title)
	if original == "" {
		return ""
	}

	cleaned := spotifyFeaturingTagPattern.ReplaceAllString(original, "")
	cleaned = spotifyTrailingFeatPattern.ReplaceAllString(cleaned, "")
	cleaned = spotifyNoiseTagPattern.ReplaceAllString(cleaned, "")
	cleaned = strings.Join(strings.Fields(cleaned), " ")
	cleaned = strings.Trim(cleaned, " -–—|[]()")
	cleaned = strings.Join(strings.Fields(cleaned), " ")
	if cleaned == "" {
		return original
	}

	return cleaned
}

func tokenNeedsRefresh(accessToken string, expiresAt time.Time) bool {
	if strings.TrimSpace(accessToken) == "" {
		return true
	}
	if expiresAt.IsZero() {
		return false
	}

	return time.Until(expiresAt) <= 5*time.Minute
}
