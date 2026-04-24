package dashboard

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/botstatus"
	"github.com/mr-cheeezz/dankbot/pkg/buildmeta"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	"github.com/mr-cheeezz/dankbot/pkg/release"
	"github.com/mr-cheeezz/dankbot/pkg/twitch/helix"
	webaccess "github.com/mr-cheeezz/dankbot/pkg/web/access"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
	"github.com/mr-cheeezz/dankbot/pkg/web/state"
)

const helixSummaryTimeout = 2 * time.Second

type handler struct {
	appState *state.State
}

type integrationAction struct {
	Kind   string `json:"kind,omitempty"`
	Label  string `json:"label"`
	Href   string `json:"href,omitempty"`
	Target string `json:"target,omitempty"`
}

type integrationSummary struct {
	ID      string              `json:"id"`
	Name    string              `json:"name"`
	Status  string              `json:"status"`
	Detail  string              `json:"detail"`
	Actions []integrationAction `json:"actions"`
}

type summaryResponse struct {
	ChannelName       string               `json:"channel_name"`
	ChannelAvatarURL  string               `json:"channel_avatar_url"`
	BotRunning        bool                 `json:"bot_running"`
	KillswitchEnabled bool                 `json:"killswitch_enabled"`
	ReleaseVersion    string               `json:"release_version"`
	WebVersion        string               `json:"web_version"`
	WebBranch         string               `json:"web_branch"`
	WebRevision       string               `json:"web_revision"`
	WebCommitTime     string               `json:"web_commit_time"`
	BotVersion        string               `json:"bot_version"`
	BotBranch         string               `json:"bot_branch"`
	BotRevision       string               `json:"bot_revision"`
	BotCommitTime     string               `json:"bot_commit_time"`
	Integrations      []integrationSummary `json:"integrations"`
}

func Register(mux *http.ServeMux, appState *state.State) {
	mux.Handle("/api/dashboard/summary", NewHandler(appState))
	mux.Handle("/api/dashboard/bot-controls", http.HandlerFunc(handler{appState: appState}.botControls))
	mux.Handle("/api/dashboard/modes", http.HandlerFunc(handler{appState: appState}.modes))
	mux.Handle("/api/dashboard/modes/settings", http.HandlerFunc(handler{appState: appState}.modesModuleSettings))
	mux.Handle("/api/dashboard/killswitch", http.HandlerFunc(handler{appState: appState}.killswitch))
	mux.Handle("/api/dashboard/audit-logs", http.HandlerFunc(handler{appState: appState}.auditLogs))
	mux.Handle("/api/dashboard/spotify", http.HandlerFunc(handler{appState: appState}.spotifyStatus))
	mux.Handle("/api/dashboard/spotify/search", http.HandlerFunc(handler{appState: appState}.spotifySearch))
	mux.Handle("/api/dashboard/spotify/queue", http.HandlerFunc(handler{appState: appState}.spotifyQueue))
	mux.Handle("/api/dashboard/spotify/playback", http.HandlerFunc(handler{appState: appState}.spotifyPlayback))
	mux.Handle("/api/dashboard/default-keywords", http.HandlerFunc(handler{appState: appState}.defaultKeywords))
	mux.Handle("/api/dashboard/modules", http.HandlerFunc(handler{appState: appState}.modulesCatalog))
	mux.Handle("/api/dashboard/modules/followers-only", http.HandlerFunc(handler{appState: appState}.followersOnlyModule))
	mux.Handle("/api/dashboard/modules/new-chatter-greeting", http.HandlerFunc(handler{appState: appState}.newChatterGreetingModule))
	mux.Handle("/api/dashboard/modules/game", http.HandlerFunc(handler{appState: appState}.gameModule))
	mux.Handle("/api/dashboard/modules/now-playing", http.HandlerFunc(handler{appState: appState}.nowPlayingModule))
	mux.Handle("/api/dashboard/modules/quotes", http.HandlerFunc(handler{appState: appState}.quoteModule))
	mux.Handle("/api/dashboard/modules/tabs", http.HandlerFunc(handler{appState: appState}.tabsModule))
	mux.Handle("/api/dashboard/modules/user-profile", http.HandlerFunc(handler{appState: appState}.userProfileModule))
	mux.Handle("/api/dashboard/modules/quotes/items", http.HandlerFunc(handler{appState: appState}.quoteModuleEntries))
	mux.Handle("/api/dashboard/modules/quotes/import", http.HandlerFunc(handler{appState: appState}.quoteModuleImport))
	mux.Handle("/api/dashboard/public-home-settings", http.HandlerFunc(handler{appState: appState}.publicHomeSettings))
	mux.Handle("/api/dashboard/alerts", http.HandlerFunc(handler{appState: appState}.alerts))
	mux.Handle("/api/dashboard/spam-filters", http.HandlerFunc(handler{appState: appState}.spamFilters))
	mux.Handle("/api/dashboard/spam-filters/hype-settings", http.HandlerFunc(handler{appState: appState}.spamFilterHypeSettings))
	mux.Handle("/api/dashboard/moderation/blocked-terms", http.HandlerFunc(handler{appState: appState}.blockedTerms))
	mux.Handle("/api/dashboard/moderation/mass-action", http.HandlerFunc(handler{appState: appState}.massModerationAction))
	mux.Handle("/api/dashboard/moderation/recent-followers", http.HandlerFunc(handler{appState: appState}.massModerationRecentFollowers))
	mux.Handle("/api/dashboard/discord-bot", http.HandlerFunc(handler{appState: appState}.discordBot))
	mux.Handle("/api/dashboard/integrations/unlink", http.HandlerFunc(handler{appState: appState}.unlinkIntegration))
	mux.Handle("/api/dashboard/roles", http.HandlerFunc(handler{appState: appState}.roles))
	mux.Handle("/api/dashboard/twitch-user-search", http.HandlerFunc(handler{appState: appState}.twitchUserSearch))
	mux.Handle("/api/dashboard/twitch-category-search", http.HandlerFunc(handler{appState: appState}.twitchCategorySearch))
}

func NewHandler(appState *state.State) http.Handler {
	h := handler{appState: appState}
	return http.HandlerFunc(h.summary)
}

func (h handler) summary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	if err := h.requireDashboardAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	response := h.buildSummary(r.Context())

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (h handler) requireDashboardAccess(r *http.Request) error {
	_, err := h.dashboardSession(r)
	return err
}

func (h handler) dashboardSession(r *http.Request) (*session.UserSession, error) {
	userSession, _, err := webaccess.LoadDashboardSession(r.Context(), r, h.appState)
	return userSession, err
}

func (h handler) requireEditorFeatureAccess(r *http.Request) (*session.UserSession, error) {
	userSession, access, err := webaccess.LoadDashboardSession(r.Context(), r, h.appState)
	if err != nil {
		return nil, err
	}
	if !webaccess.CanAccessEditorFeatures(access) {
		return nil, errors.New("dashboard access denied")
	}
	return userSession, nil
}

func (h handler) requireIntegrationsAccess(r *http.Request) (*session.UserSession, error) {
	userSession, access, err := webaccess.LoadDashboardSession(r.Context(), r, h.appState)
	if err != nil {
		return nil, err
	}
	if !webaccess.CanManageIntegrations(access) {
		return nil, errors.New("dashboard access denied")
	}
	return userSession, nil
}

func (h handler) buildSummary(ctx context.Context) summaryResponse {
	summary := summaryResponse{
		ChannelName:      "",
		ChannelAvatarURL: "",
		ReleaseVersion:   strings.TrimSpace(release.Current),
		Integrations: []integrationSummary{
			{
				ID:     "twitch",
				Name:   "Twitch",
				Status: "unlinked",
				Detail: "streamer + bot auth routes ready",
				Actions: []integrationAction{
					{Kind: "navigate", Label: "link streamer", Href: "/auth/streamer", Target: "streamer"},
					{Kind: "navigate", Label: "link bot", Href: "/auth/bot", Target: "bot"},
				},
			},
			{
				ID:      "spotify",
				Name:    "Spotify",
				Status:  "unlinked",
				Detail:  "playback, queue, search, recent history",
				Actions: []integrationAction{{Kind: "navigate", Label: "link spotify", Href: "/auth/spotify"}},
			},
			{
				ID:      "roblox",
				Name:    "Roblox",
				Status:  "unlinked",
				Detail:  "oauth plus cookie-backed presence and groups",
				Actions: []integrationAction{{Kind: "navigate", Label: "link roblox", Href: "/auth/roblox"}},
			},
			{
				ID:     "discord",
				Name:   "Discord Bot",
				Status: "available",
				Detail: "discord install route can be enabled from config",
			},
			{
				ID:      "streamelements",
				Name:    "StreamElements",
				Status:  "unlinked",
				Detail:  "oauth linking for alerts and realtime events",
				Actions: []integrationAction{{Kind: "navigate", Label: "link streamelements", Href: "/auth/streamelements"}},
			},
			{
				ID:      "streamlabs",
				Name:    "Streamlabs",
				Status:  "unlinked",
				Detail:  "oauth linking plus socket token for alerts",
				Actions: []integrationAction{{Kind: "navigate", Label: "link streamlabs", Href: "/auth/streamlabs"}},
			},
		},
	}
	webInfo := buildmeta.Detect(release.Current)
	summary.WebVersion = strings.TrimSpace(webInfo.Version)
	summary.WebBranch = strings.TrimSpace(webInfo.Branch)
	summary.WebRevision = strings.TrimSpace(webInfo.Revision)
	summary.WebCommitTime = strings.TrimSpace(webInfo.CommitTime)

	if h.appState == nil || h.appState.Config == nil {
		return summary
	}

	if !h.appState.Config.StreamElements.Enabled {
		summary.Integrations[4].Status = "disabled"
		summary.Integrations[4].Detail = "streamelements oauth is not enabled in config"
		summary.Integrations[4].Actions = nil
	}

	if !h.appState.Config.Streamlabs.Enabled {
		summary.Integrations[5].Status = "disabled"
		summary.Integrations[5].Detail = "streamlabs oauth is not enabled in config"
		summary.Integrations[5].Actions = nil
	}

	if h.appState.Redis != nil {
		if payload, err := h.appState.Redis.Get(ctx, botstatus.RedisKey); err == nil {
			if heartbeat, err := botstatus.Unmarshal(payload); err == nil && heartbeat != nil {
				summary.BotRunning = true
				summary.BotVersion = strings.TrimSpace(heartbeat.Version)
				summary.BotBranch = strings.TrimSpace(heartbeat.Branch)
				summary.BotRevision = strings.TrimSpace(heartbeat.Revision)
				summary.BotCommitTime = strings.TrimSpace(heartbeat.CommitTime)
			}
		}
	}

	if h.appState.Postgres != nil {
		stateStore := postgres.NewBotStateStore(h.appState.Postgres)
		if botState, err := stateStore.Get(ctx); err == nil && botState != nil {
			summary.KillswitchEnabled = botState.KillswitchEnabled
		}
	}

	summary.ChannelName = strings.TrimSpace(h.appState.Config.Main.StreamerID)

	if streamerAccount, _ := h.appState.TwitchAccounts.Get(ctx, postgres.TwitchAccountKindStreamer); streamerAccount != nil {
		name := strings.TrimSpace(streamerAccount.DisplayName)
		if name == "" {
			name = strings.TrimSpace(streamerAccount.Login)
		}
		if name != "" {
			summary.ChannelName = name
		}
	}

	streamerLinked := false
	botLinked := false
	if streamerAccount, _ := h.appState.TwitchAccounts.Get(ctx, postgres.TwitchAccountKindStreamer); streamerAccount != nil {
		streamerLinked = true
	}
	if botAccount, _ := h.appState.TwitchAccounts.Get(ctx, postgres.TwitchAccountKindBot); botAccount != nil {
		botLinked = true
	}

	switch {
	case streamerLinked && botLinked:
		summary.Integrations[0].Status = "linked"
		summary.Integrations[0].Detail = "streamer + bot connected"
		summary.Integrations[0].Actions = []integrationAction{
			{Kind: "unlink", Label: "unlink streamer", Target: "streamer"},
			{Kind: "unlink", Label: "unlink bot", Target: "bot"},
		}
	case streamerLinked:
		summary.Integrations[0].Status = "partial"
		summary.Integrations[0].Detail = "streamer connected, bot still needs linking"
		summary.Integrations[0].Actions = []integrationAction{
			{Kind: "unlink", Label: "unlink streamer", Target: "streamer"},
			{Kind: "navigate", Label: "link bot", Href: "/auth/bot", Target: "bot"},
		}
	case botLinked:
		summary.Integrations[0].Status = "partial"
		summary.Integrations[0].Detail = "bot connected, streamer still needs linking"
		summary.Integrations[0].Actions = []integrationAction{
			{Kind: "navigate", Label: "link streamer", Href: "/auth/streamer", Target: "streamer"},
			{Kind: "unlink", Label: "unlink bot", Target: "bot"},
		}
	}

	if spotifyAccount, _ := h.appState.SpotifyAccounts.Get(ctx, postgres.SpotifyAccountKindStreamer); spotifyAccount != nil {
		summary.Integrations[1].Status = "linked"
		summary.Integrations[1].Actions = []integrationAction{{Kind: "unlink", Label: "unlink spotify"}}
		if displayName := strings.TrimSpace(spotifyAccount.DisplayName); displayName != "" {
			summary.Integrations[1].Detail = "connected as " + displayName
		}
	}

	if robloxAccount, _ := h.appState.RobloxAccounts.Get(ctx, postgres.RobloxAccountKindStreamer); robloxAccount != nil {
		summary.Integrations[2].Status = "linked"
		summary.Integrations[2].Actions = []integrationAction{{Kind: "unlink", Label: "unlink roblox"}}
		displayName := strings.TrimSpace(robloxAccount.DisplayName)
		if displayName == "" {
			displayName = strings.TrimSpace(robloxAccount.Username)
		}
		if displayName != "" {
			summary.Integrations[2].Detail = "connected as " + displayName
		}
	}

	if h.appState.Config.Discord.Enabled && h.appState.DiscordOAuth != nil {
		discordInstallation, _ := h.appState.DiscordBotInstallation.Get(ctx)
		botTokenConfigured := strings.TrimSpace(h.appState.Config.Discord.BotToken) != "" &&
			!strings.EqualFold(strings.TrimSpace(h.appState.Config.Discord.BotToken), "your_discord_bot_token")
		if discordInstallation != nil && strings.TrimSpace(discordInstallation.GuildID) != "" {
			summary.Integrations[3].Status = "linked"
			summary.Integrations[3].Detail = "installed in guild " + strings.TrimSpace(discordInstallation.GuildID)
			if botTokenConfigured {
				summary.Integrations[3].Detail += " and runtime token is configured"
			} else {
				summary.Integrations[3].Detail += "; runtime token still needs to be configured"
			}
			summary.Integrations[3].Actions = []integrationAction{
				{Kind: "navigate", Label: "reinstall bot", Href: "/auth/discord"},
				{Kind: "unlink", Label: "unlink discord"},
			}
		} else {
			summary.Integrations[3].Status = "available"
			if botTokenConfigured {
				summary.Integrations[3].Detail = "runtime token configured, install route ready"
			} else {
				summary.Integrations[3].Detail = "discord install route ready"
			}
			summary.Integrations[3].Actions = []integrationAction{{Kind: "navigate", Label: "install bot", Href: "/auth/discord"}}
		}
	} else {
		summary.Integrations[3].Status = "disabled"
		summary.Integrations[3].Detail = "discord oauth is not enabled in config"
		summary.Integrations[3].Actions = nil
	}

	if streamElementsAccount, _ := h.appState.StreamElementsAccounts.Get(ctx, postgres.StreamElementsAccountKindStreamer); streamElementsAccount != nil {
		summary.Integrations[4].Status = "linked"
		summary.Integrations[4].Actions = []integrationAction{{Kind: "unlink", Label: "unlink streamelements"}}
		detail := strings.TrimSpace(streamElementsAccount.DisplayName)
		if detail == "" {
			detail = strings.TrimSpace(streamElementsAccount.Username)
		}
		if detail != "" {
			summary.Integrations[4].Detail = "connected as " + detail
		} else {
			summary.Integrations[4].Detail = "oauth account linked"
		}
	}

	if streamlabsAccount, _ := h.appState.StreamlabsAccounts.Get(ctx, postgres.StreamlabsAccountKindStreamer); streamlabsAccount != nil {
		summary.Integrations[5].Status = "linked"
		summary.Integrations[5].Actions = []integrationAction{{Kind: "unlink", Label: "unlink streamlabs"}}
		if displayName := strings.TrimSpace(streamlabsAccount.DisplayName); displayName != "" {
			summary.Integrations[5].Detail = "connected as " + displayName
		} else {
			summary.Integrations[5].Detail = "oauth account linked"
		}
	}

	enrichSummaryWithTwitchUser(ctx, h.appState, &summary)

	return summary
}

func enrichSummaryWithTwitchUser(ctx context.Context, appState *state.State, summary *summaryResponse) {
	if appState == nil || appState.Config == nil || summary == nil {
		return
	}

	streamerID := strings.TrimSpace(appState.Config.Main.StreamerID)
	if streamerID == "" {
		return
	}

	tokenCtx, cancel := context.WithTimeout(ctx, helixSummaryTimeout)
	defer cancel()

	appToken, err := appState.TwitchOAuth.AppToken(tokenCtx)
	if err != nil {
		return
	}

	client := helix.NewClient(appState.Config.Twitch.ClientID, appToken.AccessToken)

	users, err := client.GetUsersByIDs(tokenCtx, []string{streamerID})
	if err != nil || len(users) == 0 {
		return
	}

	user := users[0]
	if displayName := strings.TrimSpace(user.DisplayName); displayName != "" {
		summary.ChannelName = displayName
	} else if login := strings.TrimSpace(user.Login); login != "" {
		summary.ChannelName = login
	}
	summary.ChannelAvatarURL = strings.TrimSpace(user.ProfileImageURL)
}

func writeMethodNotAllowed(w http.ResponseWriter, allowedMethod string) {
	w.Header().Set("Allow", allowedMethod)
	w.WriteHeader(http.StatusMethodNotAllowed)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
}
