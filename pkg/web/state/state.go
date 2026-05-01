package state

import (
	"strings"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/common/config"
	discordoauth "github.com/mr-cheeezz/dankbot/pkg/discord/oauth"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	redispkg "github.com/mr-cheeezz/dankbot/pkg/redis"
	robloxoauth "github.com/mr-cheeezz/dankbot/pkg/roblox/oauth"
	spotifyoauth "github.com/mr-cheeezz/dankbot/pkg/spotify/oauth"
	streamelementsoauth "github.com/mr-cheeezz/dankbot/pkg/streamelements/oauth"
	streamlabsoauth "github.com/mr-cheeezz/dankbot/pkg/streamlabs/oauth"
	"github.com/mr-cheeezz/dankbot/pkg/twitch/eventsub"
	twitchoauth "github.com/mr-cheeezz/dankbot/pkg/twitch/oauth"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
)

type State struct {
	Config                 *config.Config
	StartedAt              time.Time
	Postgres               *postgres.Client
	Redis                  *redispkg.Client
	DiscordOAuth           *discordoauth.Service
	DiscordBotInstallation *postgres.DiscordBotInstallationStore
	DiscordBotSettings     *postgres.DiscordBotSettingsStore
	DiscordLogSettings     *postgres.DiscordLogSettingsStore
	TwitchOAuth            *twitchoauth.Service
	TwitchAccounts         *postgres.TwitchAccountStore
	SpotifyOAuth           *spotifyoauth.Service
	SpotifyAccounts        *postgres.SpotifyAccountStore
	RobloxOAuth            *robloxoauth.Service
	RobloxAccounts         *postgres.RobloxAccountStore
	StreamlabsOAuth        *streamlabsoauth.Service
	StreamlabsAccounts     *postgres.StreamlabsAccountStore
	StreamElementsOAuth    *streamelementsoauth.Service
	StreamElementsAccounts *postgres.StreamElementsAccountStore
	DashboardRoles         *postgres.DashboardRoleStore
	DefaultKeywords        *postgres.DefaultKeywordSettingStore
	FollowersOnlyModule    *postgres.FollowersOnlyModuleSettingsStore
	NewChatterGreeting     *postgres.NewChatterGreetingModuleSettingsStore
	GameModule             *postgres.GameModuleSettingsStore
	NowPlayingModule       *postgres.NowPlayingModuleSettingsStore
	QuoteModule            *postgres.QuoteModuleSettingsStore
	RustLogModule          *postgres.RustLogModuleSettingsStore
	TabsModule             *postgres.TabsModuleSettingsStore
	UserProfileModule      *postgres.UserProfileModuleSettingsStore
	ModesModule            *postgres.ModesModuleSettingsStore
	ModuleCatalog          *postgres.ModuleCatalogStore
	PublicHomeSettings     *postgres.PublicHomeSettingsStore
	SpamFilters            *postgres.SpamFilterStore
	SpamFilterHypeSettings *postgres.SpamFilterHypeSettingsStore
	BlockedTerms           *postgres.BlockedTermStore
	AlertSettings          *postgres.AlertSettingsStore
	AuditLogs              *postgres.AuditLogStore
	Sessions               *session.Store
	EventSubSubscriptions  *postgres.EventSubSubscriptionStore
	EventSubActivity       *postgres.EventSubActivityStore
	EventSub               *eventsub.Service
}

func New(cfg *config.Config, postgresClient *postgres.Client, redisClient *redispkg.Client) *State {
	siteLoginRedirectURI := strings.TrimSpace(cfg.Twitch.RedirectURI)
	if siteLoginRedirectURI == "" {
		siteLoginRedirectURI = strings.TrimRight(cfg.Web.PublicURL, "/") + "/connected"
	}

	twitchConnectRedirectURI := strings.TrimSpace(cfg.Twitch.ConnectRedirectURI)
	if twitchConnectRedirectURI == "" {
		twitchConnectRedirectURI = strings.TrimRight(cfg.Web.PublicURL, "/") + "/d/authorized"
	}

	robloxRedirectURI := strings.TrimSpace(cfg.Roblox.RedirectURI)
	if robloxRedirectURI == "" {
		robloxRedirectURI = strings.TrimRight(cfg.Web.PublicURL, "/") + "/d/authorized"
	}

	spotifyRedirectURI := strings.TrimSpace(cfg.Spotify.RedirectURI)
	if spotifyRedirectURI == "" {
		spotifyRedirectURI = strings.TrimRight(cfg.Web.PublicURL, "/") + "/d/authorized"
	}

	discordRedirectURI := strings.TrimSpace(cfg.Discord.RedirectURI)
	if discordRedirectURI == "" {
		discordRedirectURI = strings.TrimRight(cfg.Web.PublicURL, "/") + "/joined"
	}

	streamlabsRedirectURI := strings.TrimSpace(cfg.Streamlabs.RedirectURI)
	if streamlabsRedirectURI == "" {
		streamlabsRedirectURI = strings.TrimRight(cfg.Web.PublicURL, "/") + "/d/authorized"
	}

	streamElementsRedirectURI := strings.TrimSpace(cfg.StreamElements.RedirectURI)
	if streamElementsRedirectURI == "" {
		streamElementsRedirectURI = strings.TrimRight(cfg.Web.PublicURL, "/") + "/d/authorized"
	}

	eventSubCallbackURL := strings.TrimSpace(cfg.TwitchEventSub.CallbackURL)
	if eventSubCallbackURL == "" {
		eventSubCallbackURL = strings.TrimRight(cfg.Web.PublicURL, "/") + "/authorized"
	}

	twitchClient := twitchoauth.NewClient(nil, cfg.Twitch.ClientID, cfg.Twitch.ClientSecret, siteLoginRedirectURI)
	twitchOAuthService := twitchoauth.NewService(twitchClient, twitchoauth.NewMemoryStateStore())
	twitchOAuthService.SetSiteLoginRedirectURI(siteLoginRedirectURI)
	twitchOAuthService.SetConnectRedirectURI(twitchConnectRedirectURI)
	discordClient := discordoauth.NewClient(nil, cfg.Discord.ClientID, cfg.Discord.ClientSecret, discordRedirectURI)
	discordOAuthService := discordoauth.NewService(discordClient, discordoauth.NewRedisStateStore(redisClient))
	discordBotInstallationStore := postgres.NewDiscordBotInstallationStore(postgresClient)
	discordBotSettingsStore := postgres.NewDiscordBotSettingsStore(postgresClient)
	discordLogSettingsStore := postgres.NewDiscordLogSettingsStore(postgresClient)
	twitchAccountStore := postgres.NewTwitchAccountStore(postgresClient)
	spotifyClient := spotifyoauth.NewClient(nil, cfg.Spotify.ClientID, cfg.Spotify.ClientSecret, spotifyRedirectURI)
	spotifyOAuthService := spotifyoauth.NewService(spotifyClient, spotifyoauth.NewRedisStateStore(redisClient))
	spotifyAccountStore := postgres.NewSpotifyAccountStore(postgresClient)
	robloxClient := robloxoauth.NewClient(nil, cfg.Roblox.ClientID, cfg.Roblox.ClientSecret, robloxRedirectURI)
	robloxOAuthService := robloxoauth.NewService(robloxClient, robloxoauth.NewRedisStateStore(redisClient))
	robloxAccountStore := postgres.NewRobloxAccountStore(postgresClient)
	streamlabsClient := streamlabsoauth.NewClient(nil, cfg.Streamlabs.ClientID, cfg.Streamlabs.ClientSecret, streamlabsRedirectURI)
	streamlabsOAuthService := streamlabsoauth.NewService(streamlabsClient, streamlabsoauth.NewRedisStateStore(redisClient))
	streamlabsAccountStore := postgres.NewStreamlabsAccountStore(postgresClient)
	streamElementsClient := streamelementsoauth.NewClient(nil, cfg.StreamElements.ClientID, cfg.StreamElements.ClientSecret, streamElementsRedirectURI)
	streamElementsOAuthService := streamelementsoauth.NewService(streamElementsClient, streamelementsoauth.NewRedisStateStore(redisClient))
	streamElementsAccountStore := postgres.NewStreamElementsAccountStore(postgresClient)
	dashboardRoleStore := postgres.NewDashboardRoleStore(postgresClient)
	defaultKeywordStore := postgres.NewDefaultKeywordSettingStore(postgresClient)
	followersOnlyModuleStore := postgres.NewFollowersOnlyModuleSettingsStore(postgresClient)
	newChatterGreetingStore := postgres.NewNewChatterGreetingModuleSettingsStore(postgresClient)
	gameModuleStore := postgres.NewGameModuleSettingsStore(postgresClient)
	nowPlayingModuleStore := postgres.NewNowPlayingModuleSettingsStore(postgresClient)
	quoteModuleStore := postgres.NewQuoteModuleSettingsStore(postgresClient)
	rustLogModuleStore := postgres.NewRustLogModuleSettingsStore(postgresClient)
	tabsModuleStore := postgres.NewTabsModuleSettingsStore(postgresClient)
	userProfileModuleStore := postgres.NewUserProfileModuleSettingsStore(postgresClient)
	modesModuleStore := postgres.NewModesModuleSettingsStore(postgresClient)
	moduleCatalogStore := postgres.NewModuleCatalogStore(postgresClient)
	publicHomeSettingsStore := postgres.NewPublicHomeSettingsStore(postgresClient)
	spamFilterStore := postgres.NewSpamFilterStore(postgresClient)
	spamFilterHypeSettingsStore := postgres.NewSpamFilterHypeSettingsStore(postgresClient)
	blockedTermStore := postgres.NewBlockedTermStore(postgresClient)
	alertSettingsStore := postgres.NewAlertSettingsStore(postgresClient)
	auditLogStore := postgres.NewAuditLogStore(postgresClient)
	eventSubSubscriptionStore := postgres.NewEventSubSubscriptionStore(postgresClient)
	eventSubActivityStore := postgres.NewEventSubActivityStore(postgresClient)

	return &State{
		Config:                 cfg,
		StartedAt:              time.Now(),
		Postgres:               postgresClient,
		Redis:                  redisClient,
		DiscordOAuth:           discordOAuthService,
		DiscordBotInstallation: discordBotInstallationStore,
		DiscordBotSettings:     discordBotSettingsStore,
		DiscordLogSettings:     discordLogSettingsStore,
		TwitchOAuth:            twitchOAuthService,
		TwitchAccounts:         twitchAccountStore,
		SpotifyOAuth:           spotifyOAuthService,
		SpotifyAccounts:        spotifyAccountStore,
		RobloxOAuth:            robloxOAuthService,
		RobloxAccounts:         robloxAccountStore,
		StreamlabsOAuth:        streamlabsOAuthService,
		StreamlabsAccounts:     streamlabsAccountStore,
		StreamElementsOAuth:    streamElementsOAuthService,
		StreamElementsAccounts: streamElementsAccountStore,
		DashboardRoles:         dashboardRoleStore,
		DefaultKeywords:        defaultKeywordStore,
		FollowersOnlyModule:    followersOnlyModuleStore,
		NewChatterGreeting:     newChatterGreetingStore,
		GameModule:             gameModuleStore,
		NowPlayingModule:       nowPlayingModuleStore,
		QuoteModule:            quoteModuleStore,
		RustLogModule:          rustLogModuleStore,
		TabsModule:             tabsModuleStore,
		UserProfileModule:      userProfileModuleStore,
		ModesModule:            modesModuleStore,
		ModuleCatalog:          moduleCatalogStore,
		PublicHomeSettings:     publicHomeSettingsStore,
		SpamFilters:            spamFilterStore,
		SpamFilterHypeSettings: spamFilterHypeSettingsStore,
		BlockedTerms:           blockedTermStore,
		AlertSettings:          alertSettingsStore,
		AuditLogs:              auditLogStore,
		Sessions:               session.NewStore(redisClient),
		EventSubSubscriptions:  eventSubSubscriptionStore,
		EventSubActivity:       eventSubActivityStore,
		EventSub: eventsub.NewService(eventsub.Config{
			Enabled:      cfg.TwitchEventSub.Enabled,
			ClientID:     cfg.Twitch.ClientID,
			Transport:    cfg.TwitchEventSub.Transport,
			Secret:       cfg.TwitchEventSub.Secret,
			CallbackURL:  eventSubCallbackURL,
			WebSocketURL: cfg.TwitchEventSub.WebSocketURL,
			SyncInterval: cfg.TwitchEventSub.SyncInterval,
			DedupeTTL:    cfg.TwitchEventSub.DedupeTTL,
		}, cfg.Main.StreamerID, twitchOAuthService, twitchAccountStore, eventSubSubscriptionStore, eventSubActivityStore, redisClient),
	}
}
