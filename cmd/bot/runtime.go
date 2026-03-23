package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/botstatus"
	"github.com/mr-cheeezz/dankbot/pkg/buildmeta"
	"github.com/mr-cheeezz/dankbot/pkg/commands"
	"github.com/mr-cheeezz/dankbot/pkg/common/config"
	discordbot "github.com/mr-cheeezz/dankbot/pkg/discord/bot"
	"github.com/mr-cheeezz/dankbot/pkg/modules"
	alertsmodule "github.com/mr-cheeezz/dankbot/pkg/modules/alerts"
	blockedtermsmodule "github.com/mr-cheeezz/dankbot/pkg/modules/blockedterms"
	defaultcommandsmodule "github.com/mr-cheeezz/dankbot/pkg/modules/defaultcommands"
	discordbotmodule "github.com/mr-cheeezz/dankbot/pkg/modules/discordbot"
	followersonlymodule "github.com/mr-cheeezz/dankbot/pkg/modules/followersonly"
	robloxmodule "github.com/mr-cheeezz/dankbot/pkg/modules/game"
	keywordsmodule "github.com/mr-cheeezz/dankbot/pkg/modules/keywords"
	modesmodule "github.com/mr-cheeezz/dankbot/pkg/modules/modes"
	newchattergreetingmodule "github.com/mr-cheeezz/dankbot/pkg/modules/newchattergreeting"
	spotifymodule "github.com/mr-cheeezz/dankbot/pkg/modules/now-playing"
	quotesmodule "github.com/mr-cheeezz/dankbot/pkg/modules/quotes"
	spamfiltersmodule "github.com/mr-cheeezz/dankbot/pkg/modules/spamfilters"
	tabsmodule "github.com/mr-cheeezz/dankbot/pkg/modules/tabs"
	"github.com/mr-cheeezz/dankbot/pkg/openai"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	redispkg "github.com/mr-cheeezz/dankbot/pkg/redis"
	spotifyoauth "github.com/mr-cheeezz/dankbot/pkg/spotify/oauth"
	"github.com/mr-cheeezz/dankbot/pkg/twitch/chat"
	twitchhelix "github.com/mr-cheeezz/dankbot/pkg/twitch/helix"
	twitchoauth "github.com/mr-cheeezz/dankbot/pkg/twitch/oauth"
)

type runtime struct {
	config             *config.Config
	postgres           *postgres.Client
	redis              *redispkg.Client
	chat               *chat.Client
	discord            *discordbot.Client
	dispatcher         *commands.Dispatcher
	modules            *modules.Runner
	accounts           *postgres.TwitchAccountStore
	spotify            *postgres.SpotifyAccountStore
	modeStore          *postgres.BotModeStore
	stateStore         *postgres.BotStateStore
	socialStore        *postgres.BotSocialPromotionStore
	publicHomeSettings *postgres.PublicHomeSettingsStore
	auditStore         *postgres.AuditLogStore
	modeModule         *modesmodule.Module
	alertsModule       *alertsmodule.Module
	discordBotModule   *discordbotmodule.Module
	greetingModule     *newchattergreetingmodule.Module
	spotifyModule      *spotifymodule.Module
	spamFiltersModule  *spamfiltersmodule.Module
	blockedTermsModule *blockedtermsmodule.Module
	chatActivityStore  *postgres.TwitchUserChatActivityStore
	chatActivityMu     sync.Mutex
	chatActivityLast   map[string]time.Time
	botAccount         *postgres.TwitchAccount
	streamer           *postgres.TwitchAccount
	onConnectOnce      sync.Once
	helixOAuth         *twitchoauth.Service
	helixClient        *twitchhelix.Client
	helixToken         *twitchoauth.Token
	helixMu            sync.Mutex
	helixSendDownUntil time.Time
	commandPrefix      string
	startedAt          time.Time
	buildInfo          buildmeta.Info
}

const helixSendTimeout = 1500 * time.Millisecond

func newRuntime(cfg *config.Config) *runtime {
	var redisClient *redispkg.Client
	if strings.TrimSpace(cfg.Redis.Addr) != "" {
		redisClient = redispkg.NewClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB, cfg.Redis.KeyPrefix)
	}
	dispatcher := commands.NewDispatcher("!")
	postgresClient := postgres.NewClient(cfg.Main.DB)
	twitchOAuthService := twitchoauth.NewService(
		twitchoauth.NewClient(nil, cfg.Twitch.ClientID, cfg.Twitch.ClientSecret, strings.TrimSpace(cfg.Twitch.RedirectURI)),
		nil,
	)
	spotifyOAuthService := spotifyoauth.NewService(
		spotifyoauth.NewClient(nil, cfg.Spotify.ClientID, cfg.Spotify.ClientSecret, strings.TrimSpace(cfg.Spotify.RedirectURI)),
		nil,
	)
	twitchAccountStore := postgres.NewTwitchAccountStore(postgresClient)
	modeStore := postgres.NewBotModeStore(postgresClient)
	stateStore := postgres.NewBotStateStore(postgresClient)
	socialStore := postgres.NewBotSocialPromotionStore(postgresClient)
	auditStore := postgres.NewAuditLogStore(postgresClient)
	chatActivityStore := postgres.NewTwitchUserChatActivityStore(postgresClient)
	publicHomeSettingsStore := postgres.NewPublicHomeSettingsStore(postgresClient)
	defaultCommandStore := postgres.NewDefaultCommandSettingStore(postgresClient)
	defaultKeywordStore := postgres.NewDefaultKeywordSettingStore(postgresClient)
	gameModuleSettingsStore := postgres.NewGameModuleSettingsStore(postgresClient)
	modeModule := modesmodule.New(modeStore, stateStore, socialStore, auditStore, cfg.Main.AdminID, cfg.Main.StreamerID)
	streamChecker := newStreamStatusChecker(cfg, twitchAccountStore)
	modeModule.SetStreamLiveChecker(streamChecker.IsLive)
	modeModule.SetTwitchTitleCoordinator(cfg.Twitch.ClientID, twitchOAuthService, twitchAccountStore)
	modeModule.SetChannelSettingsStore(publicHomeSettingsStore)
	modeModule.SetModesModuleSettingsStore(postgres.NewModesModuleSettingsStore(postgresClient))
	alertsModule := alertsmodule.New(redisClient, stateStore, cfg.Main.StreamerID)
	discordModule := discordbotmodule.New(
		postgres.NewDiscordBotSettingsStore(postgresClient),
		stateStore,
		"",
		cfg.Main.AdminID,
	)
	discordModule.SetStreamLiveChecker(streamChecker.IsLive)
	discordModule.SetStreamGameChecker(streamChecker.LiveAndGame)
	spotifyModule := spotifymodule.New(
		postgres.NewSpotifyAccountStore(postgresClient),
		twitchAccountStore,
		auditStore,
		spotifyOAuthService,
		cfg.Main.AdminID,
		cfg.Main.StreamerID,
	)
	spotifyModule.SetStreamLiveChecker(streamChecker.IsLive)
	spamFiltersModule := spamfiltersmodule.New(postgres.NewSpamFilterStore(postgresClient))
	blockedTermsModule := blockedtermsmodule.New(postgres.NewBlockedTermStore(postgresClient))
	keywordsModule := keywordsmodule.New(
		postgres.NewKeywordStore(postgresClient),
		defaultKeywordStore,
		stateStore,
		twitchAccountStore,
		cfg.Main.AdminID,
		cfg.Main.StreamerID,
	)
	keywordsModule.SetModeStore(modeStore)
	keywordsModule.SetGameModuleSettingsStore(gameModuleSettingsStore)
	keywordsModule.SetNowPlayingModuleSettingsStore(postgres.NewNowPlayingModuleSettingsStore(postgresClient))
	if cfg.OpenAI.Enabled && cfg.OpenAI.KeywordValidation {
		keywordsModule.SetSemanticValidator(openai.NewClient(cfg.OpenAI.APIKey, cfg.OpenAI.Model, cfg.OpenAI.Timeout))
	}
	moduleRunner := modules.NewRunner(dispatcher)
	moduleRunner.Register(defaultcommandsmodule.New(time.Now().UTC(), botVersion, defaultCommandStore))
	moduleRunner.Register(alertsModule)
	moduleRunner.Register(discordModule)
	moduleRunner.Register(followersonlymodule.New(
		postgres.NewFollowersOnlyModuleSettingsStore(postgresClient),
		twitchAccountStore,
		twitchOAuthService,
		cfg.Twitch.ClientID,
		cfg.Main.StreamerID,
		cfg.Main.BotID,
	))
	moduleRunner.Register(modeModule)
	moduleRunner.Register(blockedTermsModule)
	moduleRunner.Register(spamFiltersModule)
	moduleRunner.Register(keywordsModule)
	greetingModule := newchattergreetingmodule.New(postgres.NewNewChatterGreetingModuleSettingsStore(postgresClient))
	moduleRunner.Register(greetingModule)
	quotesModule := quotesmodule.New(postgres.NewQuoteStore(postgresClient), auditStore, cfg.Main.AdminID, cfg.Main.StreamerID)
	quotesModule.SetSettingsStore(postgres.NewQuoteModuleSettingsStore(postgresClient))
	moduleRunner.Register(quotesModule)
	moduleRunner.Register(tabsmodule.New(
		postgres.NewUserTabStore(postgresClient),
		postgres.NewTabsModuleSettingsStore(postgresClient),
		cfg.Main.AdminID,
		cfg.Main.StreamerID,
	))
	moduleRunner.Register(robloxmodule.New(
		cfg.Roblox.Cookie,
		cfg.Twitch.ClientID,
		cfg.Main.StreamerID,
		postgres.NewRobloxPlaytimeStore(postgresClient),
		gameModuleSettingsStore,
		twitchOAuthService,
		twitchAccountStore,
	))
	moduleRunner.Register(spotifyModule)

	return &runtime{
		config:             cfg,
		postgres:           postgresClient,
		redis:              redisClient,
		dispatcher:         dispatcher,
		modules:            moduleRunner,
		accounts:           twitchAccountStore,
		spotify:            postgres.NewSpotifyAccountStore(postgresClient),
		modeStore:          modeStore,
		stateStore:         stateStore,
		socialStore:        socialStore,
		publicHomeSettings: publicHomeSettingsStore,
		auditStore:         auditStore,
		modeModule:         modeModule,
		alertsModule:       alertsModule,
		discordBotModule:   discordModule,
		greetingModule:     greetingModule,
		spotifyModule:      spotifyModule,
		spamFiltersModule:  spamFiltersModule,
		blockedTermsModule: blockedTermsModule,
		chatActivityStore:  chatActivityStore,
		chatActivityLast:   make(map[string]time.Time),
		helixOAuth:         twitchOAuthService,
		commandPrefix:      "!",
		startedAt:          time.Now().UTC(),
		buildInfo:          buildmeta.Detect(botVersion),
	}
}

func (r *runtime) Run(ctx context.Context) error {
	if r.redis != nil {
		defer r.redis.Close()
	}
	defer r.postgres.Close()

	if err := r.modules.Start(ctx); err != nil {
		return err
	}
	if err := r.configureCommandPrefix(ctx); err != nil {
		fmt.Printf("warning: could not load command prefix from channel settings: %v\n", err)
	}

	if err := r.initializeChat(ctx); err != nil {
		return err
	}
	if err := r.initializeDiscord(ctx); err != nil {
		return err
	}

	go r.runTwitchTokenRefresh(ctx)
	go r.runBotHeartbeat(ctx)

	if err := r.chat.Start(ctx); err != nil {
		return err
	}
	if r.discord != nil {
		if err := r.discord.Start(ctx); err != nil {
			fmt.Printf("warning: discord runtime disabled: %v\n", err)
			_ = r.discord.Stop(context.Background())
			r.discord = nil
		}
	}

	select {
	case <-ctx.Done():
		r.clearBotHeartbeat(context.Background())
		r.sendConfiguredLifecycleMessage(r.config.UpDown.Down)
		if r.discord != nil {
			_ = r.discord.Stop(context.Background())
		}
		return r.chat.Stop(context.Background())
	case err := <-r.chat.Errors():
		if err == nil {
			return nil
		}
		return err
	case err, ok := <-r.discordErrors():
		if !ok || err == nil {
			return nil
		}
		return err
	}
}

func (r *runtime) configureCommandPrefix(ctx context.Context) error {
	prefix := "!"
	if r.publicHomeSettings != nil {
		if err := r.publicHomeSettings.EnsureDefault(ctx); err != nil {
			return err
		}

		settings, err := r.publicHomeSettings.Get(ctx)
		if err != nil {
			return err
		}
		if settings != nil {
			prefix = settings.CommandPrefix
		}
	}

	prefix = normalizeRuntimeCommandPrefix(prefix)
	r.commandPrefix = prefix
	if r.dispatcher != nil {
		r.dispatcher.SetPrefix(prefix)
	}

	return nil
}

func (r *runtime) runBotHeartbeat(ctx context.Context) {
	if r.redis == nil {
		return
	}

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		if err := r.writeBotHeartbeat(ctx); err != nil {
			fmt.Printf("bot heartbeat error: %v\n", err)
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (r *runtime) writeBotHeartbeat(ctx context.Context) error {
	if r.redis == nil {
		return nil
	}

	payload, err := botstatus.Heartbeat{
		StartedAt:     r.startedAt.UTC(),
		LastSeenAt:    time.Now().UTC(),
		BotLogin:      twitchAccountLogin(r.botAccount),
		StreamerLogin: twitchAccountLogin(r.streamer),
		Version:       strings.TrimSpace(r.buildInfo.Version),
		Branch:        strings.TrimSpace(r.buildInfo.Branch),
		Revision:      strings.TrimSpace(r.buildInfo.Revision),
		CommitTime:    strings.TrimSpace(r.buildInfo.CommitTime),
	}.Marshal()
	if err != nil {
		return err
	}

	return r.redis.Set(ctx, botstatus.RedisKey, payload, 45*time.Second)
}

func (r *runtime) clearBotHeartbeat(ctx context.Context) {
	if r.redis == nil {
		return
	}

	_ = r.redis.Delete(ctx, botstatus.RedisKey)
}

func twitchAccountLogin(account *postgres.TwitchAccount) string {
	if account == nil {
		return ""
	}

	return strings.TrimSpace(account.Login)
}

func (r *runtime) initializeChat(ctx context.Context) error {
	if err := r.refreshTwitchAccounts(ctx); err != nil {
		return err
	}
	botAccount := r.botAccount
	if botAccount == nil {
		return fmt.Errorf("twitch bot account is not linked yet")
	}
	if r.config.Main.BotID != "" && botAccount.TwitchUserID != r.config.Main.BotID {
		return fmt.Errorf("linked twitch bot account %s does not match config bot id %s", botAccount.TwitchUserID, r.config.Main.BotID)
	}

	streamerAccount := r.streamer
	if streamerAccount == nil {
		resolvedStreamer, err := r.resolveConfiguredStreamer(ctx)
		if err != nil {
			return err
		}
		streamerAccount = resolvedStreamer
	}
	if streamerAccount == nil {
		return fmt.Errorf("twitch streamer account is not linked yet and could not be resolved from config")
	}
	if r.config.Main.StreamerID != "" && streamerAccount.TwitchUserID != r.config.Main.StreamerID {
		return fmt.Errorf("linked twitch streamer account %s does not match config streamer id %s", streamerAccount.TwitchUserID, r.config.Main.StreamerID)
	}

	if r.usesHelixChatSend() {
		required := missingScopes(botAccount.Scopes, "chat:read", "user:bot", "user:write:chat")
		if len(required) > 0 {
			return fmt.Errorf("linked twitch bot account is missing required helix chat scopes: %s", strings.Join(required, ", "))
		}
	} else {
		required := missingScopes(botAccount.Scopes, "chat:read", "chat:edit")
		if len(required) > 0 {
			return fmt.Errorf("linked twitch bot account is missing required irc chat scopes: %s", strings.Join(required, ", "))
		}
	}
	if r.usesHelixChatSend() {
		if botAccount.TwitchUserID == streamerAccount.TwitchUserID {
			fmt.Println("warning: helix chat send is enabled, but bot and streamer are the same twitch account, so a separate bot badge will not appear")
		} else if len(missingScopes(streamerAccount.Scopes, "channel:bot")) > 0 {
			fmt.Println("warning: helix chat send is enabled, but the streamer account is missing channel:bot; bot-badge sends may fail unless the bot is a moderator")
		}
	}

	r.botAccount = botAccount
	r.streamer = streamerAccount
	r.chat = chat.NewClient(chat.Config{
		BotLogin:   botAccount.Login,
		OAuthToken: ircToken(botAccount.AccessToken),
		Channels:   []string{streamerAccount.Login},
	})

	r.chat.RegisterHandlers(chat.Handlers{
		OnConnect: func() {
			r.onConnectOnce.Do(func() {
				go func() {
					time.Sleep(1500 * time.Millisecond)
					r.sendConfiguredLifecycleMessage(r.config.UpDown.Up)
				}()
			})
		},
		OnPrivateMessage: r.onPrivateMessage,
		OnNotice: func(message chat.NoticeMessage) {
			fmt.Printf("[notice:%s] %s\n", message.Channel, message.Text)
		},
		OnReconnect: func() {
			fmt.Println("twitch requested reconnect")
		},
	})

	if r.modeModule != nil {
		r.modeModule.SetChatOutput(streamerAccount.Login, r.sendChannelMessage)
	}
	if r.alertsModule != nil {
		r.alertsModule.SetChatOutput(streamerAccount.Login, r.sendChannelMessage)
	}
	if r.spotifyModule != nil {
		r.spotifyModule.SetChatOutput(streamerAccount.Login, r.sendChannelMessage)
	}
	if r.greetingModule != nil {
		r.greetingModule.SetChatOutput(streamerAccount.Login, r.sendChannelMessage)
	}
	if r.spamFiltersModule != nil {
		r.spamFiltersModule.SetModerationActions(r.deleteChatMessage, r.timeoutChatUser)
	}
	if r.blockedTermsModule != nil {
		r.blockedTermsModule.SetModerationActions(
			r.deleteChatMessage,
			r.timeoutChatUser,
			r.warnChatUser,
			r.banChatUser,
		)
	}

	return nil
}

func (r *runtime) resolveConfiguredStreamer(ctx context.Context) (*postgres.TwitchAccount, error) {
	streamerID := strings.TrimSpace(r.config.Main.StreamerID)
	if streamerID == "" {
		return nil, fmt.Errorf("twitch streamer account is not linked yet and main.streamer_id is empty")
	}

	if r.helixOAuth == nil {
		return nil, fmt.Errorf("twitch streamer account is not linked yet and twitch oauth service is not configured")
	}

	token, err := r.helixOAuth.AppToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolve streamer from config via twitch app token: %w", err)
	}
	if token == nil || strings.TrimSpace(token.AccessToken) == "" {
		return nil, fmt.Errorf("resolve streamer from config via twitch app token: empty access token")
	}

	client := twitchhelix.NewClientWithHTTPClient(
		&http.Client{Timeout: 2 * time.Second},
		r.config.Twitch.ClientID,
		token.AccessToken,
	)

	users, err := client.GetUsersByIDs(ctx, []string{streamerID})
	if err != nil {
		return nil, fmt.Errorf("resolve streamer from config via twitch helix: %w", err)
	}
	if len(users) == 0 {
		return nil, fmt.Errorf("resolve streamer from config via twitch helix: streamer id %s was not found", streamerID)
	}

	resolved := &postgres.TwitchAccount{
		Kind:         postgres.TwitchAccountKindStreamer,
		TwitchUserID: strings.TrimSpace(users[0].ID),
		Login:        strings.TrimSpace(users[0].Login),
		DisplayName:  strings.TrimSpace(users[0].DisplayName),
	}
	r.streamer = resolved

	return resolved, nil
}

func (r *runtime) initializeDiscord(ctx context.Context) error {
	_ = ctx
	if !r.config.Discord.Enabled {
		return nil
	}
	token := strings.TrimSpace(r.config.Discord.BotToken)
	if token == "" || strings.EqualFold(token, "your_discord_bot_token") {
		fmt.Println("warning: discord is enabled but discord.bot_token is not configured; skipping discord runtime")
		return nil
	}

	client, err := discordbot.NewClient(discordbot.Config{
		BotToken:  token,
		ModRoleID: r.config.Discord.ModRoleID,
	})
	if err != nil {
		fmt.Printf("warning: discord runtime disabled: %v\n", err)
		return nil
	}

	client.RegisterHandlers(discordbot.Handlers{
		OnMessage: r.onDiscordMessage,
	})
	r.discord = client
	if r.discordBotModule != nil {
		r.discordBotModule.SetDiscordSender(client.SendMessage)
		r.discordBotModule.SetDiscordEmbedSender(client.SendEmbed)
		r.discordBotModule.SetDiscordRichEmbedSender(client.SendEmbedWithComponents)
	}

	return nil
}

func (r *runtime) runTwitchTokenRefresh(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.refreshTwitchAccounts(ctx); err != nil {
				fmt.Printf("twitch token refresh error: %v\n", err)
			}
		}
	}
}

func (r *runtime) refreshTwitchAccounts(ctx context.Context) error {
	botAccount, tokenChanged, err := r.refreshTwitchAccount(ctx, postgres.TwitchAccountKindBot)
	if err != nil {
		return err
	}
	if botAccount != nil {
		r.botAccount = botAccount
		if tokenChanged && r.chat != nil {
			r.chat.SetOAuthToken(ircToken(botAccount.AccessToken))
		}
	}

	streamerAccount, _, err := r.refreshTwitchAccount(ctx, postgres.TwitchAccountKindStreamer)
	if err != nil {
		return err
	}
	if streamerAccount != nil {
		r.streamer = streamerAccount
	}

	return nil
}

func (r *runtime) refreshTwitchAccount(ctx context.Context, kind postgres.TwitchAccountKind) (*postgres.TwitchAccount, bool, error) {
	account, err := r.accounts.Get(ctx, kind)
	if err != nil || account == nil {
		return account, false, err
	}

	if r.helixOAuth == nil {
		return account, false, nil
	}

	needsRefresh := oauthTokenNeedsRefresh(account.AccessToken, account.ExpiresAt)
	if !needsRefresh {
		if validation, err := r.helixOAuth.ValidateToken(ctx, account.AccessToken); err == nil && validation != nil {
			return account, false, nil
		} else if strings.TrimSpace(account.RefreshToken) == "" {
			return nil, false, fmt.Errorf("validate twitch %s token: %w", kind, err)
		}
	}

	if strings.TrimSpace(account.RefreshToken) == "" {
		return account, false, nil
	}

	token, err := r.helixOAuth.RefreshToken(ctx, account.RefreshToken)
	if err != nil {
		return nil, false, fmt.Errorf("refresh twitch %s token: %w", kind, err)
	}

	validation, err := r.helixOAuth.ValidateToken(ctx, token.AccessToken)
	if err != nil {
		return nil, false, fmt.Errorf("validate refreshed twitch %s token: %w", kind, err)
	}

	previousAccessToken := account.AccessToken
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

	if err := r.accounts.Save(ctx, *account); err != nil {
		return nil, false, err
	}

	return account, account.AccessToken != previousAccessToken, nil
}

func (r *runtime) onPrivateMessage(message chat.Message) {
	if r.botAccount != nil && message.SenderID == r.botAccount.TwitchUserID {
		return
	}
	r.trackChatActivity(message)

	if r.killswitchEnabled(message.Text) {
		return
	}

	if r.modules != nil {
		result, err := r.modules.HandleMessage(modules.CommandContext{
			Platform:      "twitch",
			Channel:       message.Channel,
			SenderID:      message.SenderID,
			Sender:        message.Sender,
			DisplayName:   message.DisplayName,
			IsModerator:   message.IsModerator,
			IsBroadcaster: message.IsBroadcaster,
			CommandPrefix: r.commandPrefix,
			FirstMessage:  message.FirstMessage,
			MessageID:     message.ReplyTo,
			Message:       message.Text,
		})
		if err != nil {
			fmt.Printf("message handler error: %v\n", err)
			return
		}
		if strings.TrimSpace(result.Reply) != "" && r.chat != nil {
			r.sendChatReply(message, result.Reply)
			return
		}
		if result.StopProcessing {
			return
		}
	}

	result, handled, err := r.dispatcher.Dispatch(commands.Context{
		Platform:      "twitch",
		Channel:       message.Channel,
		SenderID:      message.SenderID,
		Sender:        message.Sender,
		DisplayName:   message.DisplayName,
		IsModerator:   message.IsModerator,
		IsBroadcaster: message.IsBroadcaster,
		Message:       message.Text,
	})
	if err != nil {
		fmt.Printf("command error: %v\n", err)
		return
	}
	if !handled || result.Reply == "" || r.chat == nil {
		return
	}

	r.sendChatReply(message, result.Reply)
}

func (r *runtime) trackChatActivity(message chat.Message) {
	if r.chatActivityStore == nil {
		return
	}
	userID := strings.TrimSpace(message.SenderID)
	userLogin := strings.TrimSpace(message.Sender)
	if userID == "" || userLogin == "" {
		return
	}

	now := time.Now().UTC()
	key := userID
	if key == "" {
		key = strings.ToLower(strings.TrimPrefix(userLogin, "@"))
	}

	r.chatActivityMu.Lock()
	lastWrite := r.chatActivityLast[key]
	if !lastWrite.IsZero() && now.Sub(lastWrite) < 30*time.Second {
		r.chatActivityMu.Unlock()
		return
	}
	r.chatActivityLast[key] = now
	r.chatActivityMu.Unlock()

	if err := r.chatActivityStore.Touch(context.Background(), userID, userLogin, message.DisplayName, now); err != nil {
		fmt.Printf("chat activity track error: %v\n", err)
	}
}

func (r *runtime) onDiscordMessage(message discordbot.Message) {
	if r.killswitchEnabled(message.Content) {
		return
	}

	if r.modules != nil {
		result, err := r.modules.HandleMessage(modules.CommandContext{
			Platform:      "discord",
			Channel:       message.ChannelID,
			SenderID:      message.SenderID,
			Sender:        message.Sender,
			DisplayName:   message.DisplayName,
			IsModerator:   message.IsModerator || message.IsOwnerOrAdmin,
			IsBroadcaster: message.IsOwnerOrAdmin,
			CommandPrefix: r.commandPrefix,
			FirstMessage:  false,
			MessageID:     "",
			Message:       message.Content,
		})
		if err != nil {
			fmt.Printf("discord message handler error: %v\n", err)
			return
		}
		if strings.TrimSpace(result.Reply) != "" && r.discord != nil {
			if err := r.discord.SendMessage(message.ChannelID, result.Reply); err != nil {
				fmt.Printf("discord send error: %v\n", err)
			}
			return
		}
		if result.StopProcessing {
			return
		}
	}

	result, handled, err := r.dispatcher.Dispatch(commands.Context{
		Platform:      "discord",
		Channel:       message.ChannelID,
		SenderID:      message.SenderID,
		Sender:        message.Sender,
		DisplayName:   message.DisplayName,
		IsModerator:   message.IsModerator || message.IsOwnerOrAdmin,
		IsBroadcaster: message.IsOwnerOrAdmin,
		Message:       message.Content,
	})
	if err != nil {
		fmt.Printf("discord command error: %v\n", err)
		return
	}
	if !handled || strings.TrimSpace(result.Reply) == "" || r.discord == nil {
		return
	}

	if err := r.discord.SendMessage(message.ChannelID, result.Reply); err != nil {
		fmt.Printf("discord send error: %v\n", err)
	}
}

func (r *runtime) killswitchEnabled(message string) bool {
	if r.stateStore == nil {
		return false
	}

	commandName := extractCommandName(message, r.commandPrefix)
	if commandName == "killswitch" || commandName == "ks" {
		return false
	}

	state, err := r.stateStore.Get(context.Background())
	if err != nil {
		fmt.Printf("bot state error: %v\n", err)
		return false
	}

	return state != nil && state.KillswitchEnabled
}

func extractCommandName(message, prefix string) string {
	prefix = normalizeRuntimeCommandPrefix(prefix)
	message = strings.TrimSpace(message)
	if !strings.HasPrefix(message, prefix) {
		return ""
	}

	commandLine := strings.TrimSpace(strings.TrimPrefix(message, prefix))
	if commandLine == "" {
		return ""
	}

	parts := strings.Fields(commandLine)
	if len(parts) == 0 {
		return ""
	}

	return strings.ToLower(parts[0])
}

func isCommandMessage(message, prefix string) bool {
	return strings.HasPrefix(strings.TrimSpace(message), normalizeRuntimeCommandPrefix(prefix))
}

func normalizeRuntimeCommandPrefix(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "!"
	}

	return trimmed
}

func (r *runtime) sendChatReply(message chat.Message, reply string) {
	_ = r.sendMessage(message.Channel, message.ReplyTo, reply)
}

func (r *runtime) sendConfiguredLifecycleMessage(template string) {
	message := strings.TrimSpace(r.renderLifecycleMessage(template))
	if message == "" {
		return
	}

	channel := ""
	if r.streamer != nil {
		channel = strings.TrimSpace(r.streamer.Login)
	}
	if channel == "" {
		cfg := r.chat.Config()
		if len(cfg.Channels) > 0 {
			channel = strings.TrimSpace(cfg.Channels[0])
		}
	}
	if channel == "" {
		return
	}

	_ = r.sendMessage(channel, "", message)
	time.Sleep(250 * time.Millisecond)
}

func (r *runtime) sendChannelMessage(channel, message string) error {
	return r.sendMessage(channel, "", message)
}

func (r *runtime) sendMessage(channel, replyTo, message string) error {
	if strings.TrimSpace(message) == "" {
		return nil
	}

	if r.usesHelixChatSend() && r.helixSendHealthy() {
		if err := r.sendMessageViaHelix(context.Background(), replyTo, message); err == nil {
			r.markHelixSendHealthy()
			return nil
		} else {
			r.markHelixSendFailure()
			fmt.Printf("helix chat send error: %v\n", err)
		}
	}

	if r.chat == nil {
		return fmt.Errorf("chat client is not configured")
	}

	if strings.TrimSpace(replyTo) != "" {
		return r.chat.Reply(channel, replyTo, message)
	}

	return r.chat.Say(channel, message)
}

func (r *runtime) sendMessageViaHelix(ctx context.Context, replyTo, message string) error {
	sendCtx, cancel := context.WithTimeout(ctx, helixSendTimeout)
	defer cancel()

	if r.botAccount == nil || r.streamer == nil {
		return fmt.Errorf("helix chat send requires linked bot and streamer accounts")
	}

	client, err := r.helixChatClient(sendCtx)
	if err != nil {
		return err
	}

	req := twitchhelix.SendChatMessageRequest{
		BroadcasterID: r.streamer.TwitchUserID,
		SenderID:      r.botAccount.TwitchUserID,
		Message:       message,
	}
	if strings.TrimSpace(replyTo) != "" {
		req.ReplyParentMessageID = strings.TrimSpace(replyTo)
	}

	result, err := client.SendChatMessage(sendCtx, req)
	if err != nil {
		return err
	}
	if result == nil {
		return fmt.Errorf("twitch returned no helix chat response")
	}
	if !result.IsSent {
		if result.DropReason != nil && strings.TrimSpace(result.DropReason.Message) != "" {
			return fmt.Errorf("twitch dropped helix chat message: %s", result.DropReason.Message)
		}
		return fmt.Errorf("twitch did not send the helix chat message")
	}

	return nil
}

func (r *runtime) helixChatClient(ctx context.Context) (*twitchhelix.Client, error) {
	r.helixMu.Lock()
	defer r.helixMu.Unlock()

	if r.helixOAuth == nil {
		return nil, fmt.Errorf("twitch oauth service is not configured")
	}

	if r.helixClient != nil && r.helixToken != nil {
		expiresAt := r.helixToken.ExpiresAt()
		if expiresAt.IsZero() || time.Until(expiresAt) > time.Minute {
			return r.helixClient, nil
		}
	}

	token, err := r.helixOAuth.AppToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch twitch app token: %w", err)
	}

	if r.helixClient == nil {
		r.helixClient = twitchhelix.NewClientWithHTTPClient(&http.Client{Timeout: helixSendTimeout}, r.config.Twitch.ClientID, token.AccessToken)
	} else {
		r.helixClient.SetAccessToken(token.AccessToken)
	}
	r.helixToken = token

	return r.helixClient, nil
}

func (r *runtime) helixSendHealthy() bool {
	r.helixMu.Lock()
	defer r.helixMu.Unlock()

	return r.helixSendDownUntil.IsZero() || time.Now().After(r.helixSendDownUntil)
}

func (r *runtime) markHelixSendFailure() {
	r.helixMu.Lock()
	defer r.helixMu.Unlock()

	r.helixSendDownUntil = time.Now().Add(30 * time.Second)
}

func (r *runtime) markHelixSendHealthy() {
	r.helixMu.Lock()
	defer r.helixMu.Unlock()

	r.helixSendDownUntil = time.Time{}
}

func (r *runtime) usesHelixChatSend() bool {
	return strings.EqualFold(strings.TrimSpace(r.config.Twitch.SendTransport), "helix")
}

func (r *runtime) discordErrors() <-chan error {
	if r.discord == nil {
		return nil
	}

	return r.discord.Errors()
}

func (r *runtime) renderLifecycleMessage(template string) string {
	template = strings.TrimSpace(template)
	if template == "" {
		return ""
	}

	replacer := strings.NewReplacer(
		"{bot}", r.botName(),
		"{ver}", botVersion,
		"{streamer}", r.streamerName(),
	)

	return replacer.Replace(template)
}

func (r *runtime) botName() string {
	if r.botAccount == nil {
		return "bot"
	}
	if name := strings.TrimSpace(r.botAccount.DisplayName); name != "" {
		return name
	}
	if name := strings.TrimSpace(r.botAccount.Login); name != "" {
		return name
	}
	if id := strings.TrimSpace(r.botAccount.TwitchUserID); id != "" {
		return id
	}

	return "bot"
}

func (r *runtime) streamerName() string {
	if r.streamer == nil {
		return "streamer"
	}
	if name := strings.TrimSpace(r.streamer.DisplayName); name != "" {
		return name
	}
	if name := strings.TrimSpace(r.streamer.Login); name != "" {
		return name
	}
	if id := strings.TrimSpace(r.streamer.TwitchUserID); id != "" {
		return id
	}

	return "streamer"
}

func ircToken(accessToken string) string {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(accessToken), "oauth:") {
		return accessToken
	}

	return "oauth:" + accessToken
}

func missingScopes(actual []string, required ...string) []string {
	actualSet := make(map[string]struct{}, len(actual))
	for _, scope := range actual {
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

func oauthTokenNeedsRefresh(accessToken string, expiresAt time.Time) bool {
	if strings.TrimSpace(accessToken) == "" {
		return true
	}
	if expiresAt.IsZero() {
		return false
	}

	return time.Until(expiresAt) <= 5*time.Minute
}

func (r *runtime) deleteChatMessage(ctx context.Context, messageID string) error {
	messageID = strings.TrimSpace(messageID)
	if messageID == "" {
		return nil
	}
	client, err := r.userHelixClient(ctx)
	if err != nil {
		return err
	}
	if r.streamer == nil || r.botAccount == nil {
		return fmt.Errorf("moderation requires linked bot and streamer accounts")
	}
	return client.DeleteChatMessages(ctx, r.streamer.TwitchUserID, r.botAccount.TwitchUserID, messageID)
}

func (r *runtime) timeoutChatUser(ctx context.Context, userID string, duration time.Duration, reason string) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil
	}
	client, err := r.userHelixClient(ctx)
	if err != nil {
		return err
	}
	if r.streamer == nil || r.botAccount == nil {
		return fmt.Errorf("moderation requires linked bot and streamer accounts")
	}

	seconds := int(duration.Round(time.Second) / time.Second)
	if seconds <= 0 {
		seconds = 30
	}

	_, err = client.BanUser(ctx, r.streamer.TwitchUserID, r.botAccount.TwitchUserID, twitchhelix.BanUserRequest{
		UserID:   userID,
		Duration: &seconds,
		Reason:   strings.TrimSpace(reason),
	})
	return err
}

func (r *runtime) warnChatUser(ctx context.Context, userID string, reason string) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil
	}
	client, err := r.userHelixClient(ctx)
	if err != nil {
		return err
	}
	if r.streamer == nil || r.botAccount == nil {
		return fmt.Errorf("moderation requires linked bot and streamer accounts")
	}

	_, err = client.WarnChatUser(ctx, r.streamer.TwitchUserID, r.botAccount.TwitchUserID, twitchhelix.WarnChatUserRequest{
		UserID: userID,
		Reason: strings.TrimSpace(reason),
	})
	return err
}

func (r *runtime) banChatUser(ctx context.Context, userID string, reason string) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil
	}
	client, err := r.userHelixClient(ctx)
	if err != nil {
		return err
	}
	if r.streamer == nil || r.botAccount == nil {
		return fmt.Errorf("moderation requires linked bot and streamer accounts")
	}

	_, err = client.BanUser(ctx, r.streamer.TwitchUserID, r.botAccount.TwitchUserID, twitchhelix.BanUserRequest{
		UserID: userID,
		Reason: strings.TrimSpace(reason),
	})
	return err
}

func (r *runtime) userHelixClient(ctx context.Context) (*twitchhelix.Client, error) {
	if r.botAccount == nil {
		return nil, fmt.Errorf("linked twitch bot account is not available")
	}
	accessToken := strings.TrimSpace(r.botAccount.AccessToken)
	if accessToken == "" {
		return nil, fmt.Errorf("linked twitch bot token is empty")
	}

	httpClient := &http.Client{Timeout: 2 * time.Second}
	return twitchhelix.NewClientWithHTTPClient(httpClient, r.config.Twitch.ClientID, accessToken), nil
}
