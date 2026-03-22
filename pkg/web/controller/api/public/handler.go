package public

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/botstatus"
	"github.com/mr-cheeezz/dankbot/pkg/buildmeta"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	"github.com/mr-cheeezz/dankbot/pkg/redis"
	robloxapi "github.com/mr-cheeezz/dankbot/pkg/roblox/api"
	spotifyapi "github.com/mr-cheeezz/dankbot/pkg/spotify/api"
	steamapi "github.com/mr-cheeezz/dankbot/pkg/steam/api"
	"github.com/mr-cheeezz/dankbot/pkg/twitch/helix"
	"github.com/mr-cheeezz/dankbot/pkg/web/state"
)

const publicSummaryTimeout = 2 * time.Second

var nonAlphaNumericSlug = regexp.MustCompile(`[^a-z0-9]+`)

type handler struct {
	appState *state.State
}

type summaryResponse struct {
	ChannelName            string              `json:"channel_name"`
	ChannelLogin           string              `json:"channel_login"`
	ChannelAvatarURL       string              `json:"channel_avatar_url"`
	StreamLive             bool                `json:"stream_live"`
	StreamTitle            string              `json:"stream_title"`
	StreamGameName         string              `json:"stream_game_name"`
	StreamStartedAt        string              `json:"stream_started_at"`
	StreamEndedAt          string              `json:"stream_ended_at"`
	ViewerCount            int                 `json:"viewer_count"`
	ChatterCount           int                 `json:"chatter_count"`
	CurrentModeKey         string              `json:"current_mode_key"`
	CurrentModeTitle       string              `json:"current_mode_title"`
	CurrentModeParam       string              `json:"current_mode_param"`
	RobloxPrivateServerURL string              `json:"roblox_private_server_url"`
	RobloxGameURL          string              `json:"roblox_game_url"`
	RobloxProfileURL       string              `json:"roblox_profile_url"`
	StreamGameURL          string              `json:"stream_game_url"`
	StreamGameSource       string              `json:"stream_game_source"`
	SteamProfileURL        string              `json:"steam_profile_url"`
	BotRunning             bool                `json:"bot_running"`
	BotStartedAt           string              `json:"bot_started_at"`
	BotLastSeenAt          string              `json:"bot_last_seen_at"`
	WebVersion             string              `json:"web_version"`
	WebBranch              string              `json:"web_branch"`
	WebRevision            string              `json:"web_revision"`
	WebCommitTime          string              `json:"web_commit_time"`
	BotVersion             string              `json:"bot_version"`
	BotBranch              string              `json:"bot_branch"`
	BotRevision            string              `json:"bot_revision"`
	BotCommitTime          string              `json:"bot_commit_time"`
	PromoLinks             []promoLinkResponse `json:"promo_links"`
	NowPlayingEnabled      bool                `json:"now_playing_enabled"`
	NowPlayingShowAlbumArt bool                `json:"now_playing_show_album_art"`
	NowPlayingShowProgress bool                `json:"now_playing_show_progress"`
	NowPlayingShowLinks    bool                `json:"now_playing_show_links"`
	NowPlayingIsPlaying    bool                `json:"now_playing_is_playing"`
	NowPlayingTrackName    string              `json:"now_playing_track_name"`
	NowPlayingAlbumName    string              `json:"now_playing_album_name"`
	NowPlayingAlbumArtURL  string              `json:"now_playing_album_art_url"`
	NowPlayingTrackURL     string              `json:"now_playing_track_url"`
	NowPlayingAlbumURL     string              `json:"now_playing_album_url"`
	NowPlayingArtistURL    string              `json:"now_playing_artist_url"`
	NowPlayingProgressMS   int                 `json:"now_playing_progress_ms"`
	NowPlayingDurationMS   int                 `json:"now_playing_duration_ms"`
	NowPlayingArtists      []string            `json:"now_playing_artists"`
}

type promoLinkResponse struct {
	Label string `json:"label"`
	Href  string `json:"href"`
}

func Register(mux *http.ServeMux, appState *state.State) {
	mux.Handle("/api/public/summary", NewHandler(appState))
	mux.Handle("/api/public/commands", http.HandlerFunc(handler{appState: appState}.commands))
	mux.Handle("/api/public/quotes", http.HandlerFunc(handler{appState: appState}.quotes))
	mux.Handle("/api/public/users/", http.HandlerFunc(handler{appState: appState}.userProfile))
}

func NewHandler(appState *state.State) http.Handler {
	return http.HandlerFunc(handler{appState: appState}.summary)
}

func (h handler) summary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	response := h.buildSummary(r.Context())

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (h handler) buildSummary(ctx context.Context) summaryResponse {
	summary := summaryResponse{
		CurrentModeKey:   "join",
		CurrentModeTitle: "join",
	}
	webInfo := buildmeta.Detect("")
	summary.WebVersion = strings.TrimSpace(webInfo.Version)
	summary.WebBranch = strings.TrimSpace(webInfo.Branch)
	summary.WebRevision = strings.TrimSpace(webInfo.Revision)
	summary.WebCommitTime = strings.TrimSpace(webInfo.CommitTime)

	if h.appState == nil || h.appState.Config == nil {
		return summary
	}

	summary.ChannelName = strings.TrimSpace(h.appState.Config.Main.StreamerID)

	if streamerAccount, _ := h.appState.TwitchAccounts.Get(ctx, postgres.TwitchAccountKindStreamer); streamerAccount != nil {
		summary.ChannelLogin = strings.TrimSpace(streamerAccount.Login)
		if name := strings.TrimSpace(streamerAccount.DisplayName); name != "" {
			summary.ChannelName = name
		} else if summary.ChannelLogin != "" {
			summary.ChannelName = summary.ChannelLogin
		}
	}

	if h.appState.Postgres != nil {
		stateStore := postgres.NewBotStateStore(h.appState.Postgres)
		modeStore := postgres.NewBotModeStore(h.appState.Postgres)

		if botState, err := stateStore.Get(ctx); err == nil && botState != nil {
			if key := strings.TrimSpace(botState.CurrentModeKey); key != "" {
				summary.CurrentModeKey = key
				summary.CurrentModeTitle = key
			}
			summary.CurrentModeParam = strings.TrimSpace(botState.CurrentModeParam)
			if mode, err := modeStore.Get(ctx, summary.CurrentModeKey); err == nil && mode != nil {
				if title := strings.TrimSpace(mode.Title); title != "" {
					summary.CurrentModeTitle = strings.ToLower(title)
				}
			}
			if summary.CurrentModeKey == "link" && looksLikeURL(summary.CurrentModeParam) {
				summary.RobloxPrivateServerURL = summary.CurrentModeParam
			}
		}
	}

	if h.appState.Redis != nil {
		if payload, err := h.appState.Redis.Get(ctx, botstatus.RedisKey); err == nil {
			if heartbeat, err := botstatus.Unmarshal(payload); err == nil && heartbeat != nil {
				summary.BotRunning = true
				summary.BotStartedAt = heartbeat.StartedAt.UTC().Format(time.RFC3339)
				summary.BotLastSeenAt = heartbeat.LastSeenAt.UTC().Format(time.RFC3339)
				summary.BotVersion = strings.TrimSpace(heartbeat.Version)
				summary.BotBranch = strings.TrimSpace(heartbeat.Branch)
				summary.BotRevision = strings.TrimSpace(heartbeat.Revision)
				summary.BotCommitTime = strings.TrimSpace(heartbeat.CommitTime)
				if summary.ChannelLogin == "" {
					summary.ChannelLogin = strings.TrimSpace(heartbeat.StreamerLogin)
				}
			}
		} else if err != nil && err != redis.ErrKeyNotFound {
			// Ignore transient Redis issues for the public page.
		}

		if startedAt, err := h.appState.Redis.Get(ctx, "eventsub:stream:started_at"); err == nil {
			summary.StreamStartedAt = strings.TrimSpace(startedAt)
		} else if err != nil && err != redis.ErrKeyNotFound {
			// Ignore transient Redis issues for the public page.
		}

		if endedAt, err := h.appState.Redis.Get(ctx, "eventsub:stream:ended_at"); err == nil {
			summary.StreamEndedAt = strings.TrimSpace(endedAt)
		} else if err != nil && err != redis.ErrKeyNotFound {
			// Ignore transient Redis issues for the public page.
		}
	}

	enrichWithTwitch(ctx, h.appState, &summary)
	enrichWithChatters(ctx, h.appState, &summary)
	enrichWithLinks(ctx, h.appState, &summary)
	enrichWithNowPlaying(ctx, h.appState, &summary)

	return summary
}

func enrichWithTwitch(ctx context.Context, appState *state.State, summary *summaryResponse) {
	if appState == nil || appState.Config == nil || appState.TwitchOAuth == nil || summary == nil {
		return
	}

	streamerID := strings.TrimSpace(appState.Config.Main.StreamerID)
	if streamerID == "" {
		return
	}

	tokenCtx, cancel := context.WithTimeout(ctx, publicSummaryTimeout)
	defer cancel()

	appToken, err := appState.TwitchOAuth.AppToken(tokenCtx)
	if err != nil {
		return
	}

	client := helix.NewClient(appState.Config.Twitch.ClientID, appToken.AccessToken)

	users, err := client.GetUsersByIDs(tokenCtx, []string{streamerID})
	if err == nil && len(users) > 0 {
		user := users[0]
		if displayName := strings.TrimSpace(user.DisplayName); displayName != "" {
			summary.ChannelName = displayName
		}
		if login := strings.TrimSpace(user.Login); login != "" {
			summary.ChannelLogin = login
		}
		summary.ChannelAvatarURL = strings.TrimSpace(user.ProfileImageURL)
	}

	streams, err := client.GetStreamsByUserIDs(tokenCtx, []string{streamerID})
	if err != nil || len(streams) == 0 {
		if summary.StreamEndedAt == "" && appState.Postgres != nil {
			playtimeStore := postgres.NewStreamGamePlaytimeStore(appState.Postgres)
			if endedAt, err := playtimeStore.LastCompletedStreamEndedAt(tokenCtx); err == nil && endedAt != nil {
				summary.StreamEndedAt = endedAt.UTC().Format(time.RFC3339)
			}
		}
		return
	}

	stream := streams[0]
	summary.StreamLive = true
	summary.StreamTitle = strings.TrimSpace(stream.Title)
	summary.StreamGameName = strings.TrimSpace(stream.GameName)
	summary.StreamStartedAt = stream.StartedAt.UTC().Format(time.RFC3339)
	summary.ViewerCount = stream.ViewerCount
	summary.StreamEndedAt = ""
}

func enrichWithChatters(ctx context.Context, appState *state.State, summary *summaryResponse) {
	if appState == nil || appState.Config == nil || appState.TwitchAccounts == nil || summary == nil {
		return
	}

	streamerID := strings.TrimSpace(appState.Config.Main.StreamerID)
	if streamerID == "" {
		return
	}

	account, err := appState.TwitchAccounts.Get(ctx, postgres.TwitchAccountKindBot)
	if err != nil || account == nil {
		return
	}

	if appState.TwitchOAuth != nil &&
		!account.ExpiresAt.IsZero() &&
		time.Until(account.ExpiresAt) <= time.Minute &&
		strings.TrimSpace(account.RefreshToken) != "" {
		token, refreshErr := appState.TwitchOAuth.RefreshToken(ctx, account.RefreshToken)
		if refreshErr == nil {
			account.AccessToken = strings.TrimSpace(token.AccessToken)
			if refreshToken := strings.TrimSpace(token.RefreshToken); refreshToken != "" {
				account.RefreshToken = refreshToken
			}
			account.TokenType = strings.TrimSpace(token.TokenType)
			account.ExpiresAt = token.ExpiresAt()
			account.LastValidatedAt = time.Now().UTC()
			if saveErr := appState.TwitchAccounts.Save(ctx, *account); saveErr != nil {
				// Ignore save errors for the public page.
			}
		}
	}

	moderatorID := strings.TrimSpace(account.TwitchUserID)
	accessToken := strings.TrimSpace(account.AccessToken)
	if moderatorID == "" || accessToken == "" {
		return
	}

	chatterCtx, cancel := context.WithTimeout(ctx, publicSummaryTimeout)
	defer cancel()

	client := helix.NewClient(appState.Config.Twitch.ClientID, accessToken)
	_, _, total, err := client.GetChatters(chatterCtx, streamerID, moderatorID, 1, "")
	if err != nil {
		return
	}

	summary.ChatterCount = total
}

func enrichWithLinks(ctx context.Context, appState *state.State, summary *summaryResponse) {
	if appState == nil || summary == nil {
		return
	}

	if strings.EqualFold(strings.TrimSpace(summary.StreamGameName), "roblox") {
		enrichWithRobloxPresence(ctx, appState, summary)
		return
	}

	gameName := strings.TrimSpace(summary.StreamGameName)
	if gameName == "" {
		return
	}

	if appState.Config != nil {
		steamUserID := strings.TrimSpace(appState.Config.Steam.UserID)
		steamAPIKey := strings.TrimSpace(appState.Config.Steam.APIKey)
		if steamUserID != "" && steamAPIKey != "" {
			steamClient := steamapi.NewClient(nil, steamAPIKey)

			steamCtx, cancel := context.WithTimeout(ctx, publicSummaryTimeout)
			profileURL, err := steamClient.ResolveProfileURL(steamCtx, steamUserID)
			cancel()
			if err == nil {
				summary.SteamProfileURL = strings.TrimSpace(profileURL)
			}

			steamCtx, cancel = context.WithTimeout(ctx, publicSummaryTimeout)
			storeURL, err := steamClient.ResolveStoreURL(steamCtx, gameName)
			cancel()
			if err == nil && strings.TrimSpace(storeURL) != "" {
				summary.StreamGameURL = strings.TrimSpace(storeURL)
				summary.StreamGameSource = "steam"
				return
			}
		}
	}

	summary.StreamGameURL = twitchCategoryURL(gameName)
	summary.StreamGameSource = "twitch"
}

func enrichWithNowPlaying(ctx context.Context, appState *state.State, summary *summaryResponse) {
	if appState == nil || summary == nil || appState.PublicHomeSettings == nil {
		return
	}

	if err := appState.PublicHomeSettings.EnsureDefault(ctx); err != nil {
		return
	}

	settings, err := appState.PublicHomeSettings.Get(ctx)
	if err != nil {
		return
	}
	if settings == nil {
		defaults := postgres.DefaultPublicHomeSettings()
		settings = &defaults
	}

	summary.NowPlayingEnabled = settings.ShowNowPlaying
	summary.NowPlayingShowAlbumArt = settings.ShowNowPlayingAlbumArt
	summary.NowPlayingShowProgress = settings.ShowNowPlayingProgress
	summary.NowPlayingShowLinks = settings.ShowNowPlayingLinks
	summary.PromoLinks = promoLinksToResponse(settings.PromoLinks)

	if !settings.ShowNowPlaying || appState.SpotifyAccounts == nil {
		return
	}

	account, err := appState.SpotifyAccounts.Get(ctx, postgres.SpotifyAccountKindStreamer)
	if err != nil || account == nil {
		return
	}

	if appState.SpotifyOAuth != nil &&
		!account.ExpiresAt.IsZero() &&
		time.Until(account.ExpiresAt) <= time.Minute &&
		strings.TrimSpace(account.RefreshToken) != "" {
		token, refreshErr := appState.SpotifyOAuth.RefreshToken(ctx, account.RefreshToken)
		if refreshErr == nil {
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
			if saveErr := appState.SpotifyAccounts.Save(ctx, *account); saveErr != nil {
				// Ignore save errors for the public page.
			}
		}
	}

	accessToken := strings.TrimSpace(account.AccessToken)
	if accessToken == "" {
		return
	}

	spotifyCtx, cancel := context.WithTimeout(ctx, publicSummaryTimeout)
	defer cancel()

	client := spotifyapi.NewClient(nil, accessToken)
	playing, err := client.GetCurrentlyPlaying(spotifyCtx, "")
	if err != nil || playing == nil || playing.Item == nil {
		return
	}

	item := playing.Item
	summary.NowPlayingIsPlaying = playing.IsPlaying
	summary.NowPlayingTrackName = strings.TrimSpace(item.Name)
	summary.NowPlayingAlbumName = strings.TrimSpace(item.Album.Name)
	summary.NowPlayingTrackURL = strings.TrimSpace(item.ExternalURLs.Spotify)
	summary.NowPlayingAlbumURL = strings.TrimSpace(item.Album.ExternalURLs.Spotify)
	summary.NowPlayingProgressMS = playing.ProgressMS
	summary.NowPlayingDurationMS = item.DurationMS
	if len(item.Album.Images) > 0 {
		summary.NowPlayingAlbumArtURL = strings.TrimSpace(item.Album.Images[0].URL)
	}

	artists := make([]string, 0, len(item.Artists))
	for index, artist := range item.Artists {
		name := strings.TrimSpace(artist.Name)
		if name != "" {
			artists = append(artists, name)
		}
		if index == 0 {
			summary.NowPlayingArtistURL = strings.TrimSpace(artist.ExternalURLs.Spotify)
		}
	}
	summary.NowPlayingArtists = artists
}

func promoLinksToResponse(items []postgres.PromoLink) []promoLinkResponse {
	out := make([]promoLinkResponse, 0, len(items))
	for _, item := range items {
		out = append(out, promoLinkResponse{
			Label: strings.TrimSpace(item.Label),
			Href:  strings.TrimSpace(item.Href),
		})
	}

	return out
}

func enrichWithRobloxPresence(ctx context.Context, appState *state.State, summary *summaryResponse) {
	if appState == nil || appState.RobloxAccounts == nil {
		return
	}

	robloxAccount, err := appState.RobloxAccounts.Get(ctx, postgres.RobloxAccountKindStreamer)
	if err != nil || robloxAccount == nil {
		return
	}

	userID, err := strconv.ParseInt(strings.TrimSpace(robloxAccount.RobloxUserID), 10, 64)
	if err != nil || userID <= 0 {
		return
	}

	summary.RobloxProfileURL = robloxProfileURL(userID)

	if strings.TrimSpace(appState.Config.Roblox.Cookie) == "" {
		return
	}

	presenceCtx, cancel := context.WithTimeout(ctx, publicSummaryTimeout)
	defer cancel()

	client := robloxapi.NewClient(nil, appState.Config.Roblox.Cookie)
	presences, err := client.GetPresences(presenceCtx, []int64{userID})
	if err != nil || len(presences) == 0 {
		return
	}

	presence := presences[0]
	rootPlaceID := presence.RootPlaceID
	if rootPlaceID == 0 {
		rootPlaceID = presence.PlaceID
	}
	if rootPlaceID > 0 {
		summary.RobloxGameURL = robloxGameURL(rootPlaceID)
	}
}

func looksLikeURL(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return false
	}

	return parsed.Scheme == "http" || parsed.Scheme == "https"
}

func robloxProfileURL(userID int64) string {
	return "https://www.roblox.com/users/" + strconv.FormatInt(userID, 10) + "/profile"
}

func robloxGameURL(rootPlaceID int64) string {
	return "https://www.roblox.com/games/" + strconv.FormatInt(rootPlaceID, 10)
}

func twitchCategoryURL(gameName string) string {
	gameName = strings.TrimSpace(strings.ToLower(gameName))
	if gameName == "" {
		return ""
	}

	slug := nonAlphaNumericSlug.ReplaceAllString(gameName, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		return ""
	}

	return "https://www.twitch.tv/directory/category/" + slug
}

func writeMethodNotAllowed(w http.ResponseWriter, allowedMethod string) {
	w.Header().Set("Allow", allowedMethod)
	w.WriteHeader(http.StatusMethodNotAllowed)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
}
