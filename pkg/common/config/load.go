package config

import (
	"fmt"
	"strings"
	"time"

	ini "gopkg.in/ini.v1"
)

func parseBool01(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func splitCommaList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}

	return out
}

func parseDuration(value string) (time.Duration, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}

	return time.ParseDuration(value)
}

func Load(path string) (*Config, error) {
	file, err := ini.Load(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	cfg := &Config{}

	mainSection := file.Section("main")
	cfg.Main = MainConfig{
		BotID:      mainSection.Key("bot_id").String(),
		StreamerID: mainSection.Key("streamer_id").String(),
		AdminID:    mainSection.Key("admin_id").String(),
		DB:         mainSection.Key("db").String(),
	}

	webSection := file.Section("web")
	bindAddr := strings.TrimSpace(webSection.Key("bind_addr").String())
	bindUnixSocket := strings.TrimSpace(webSection.Key("bind_unix_socket").String())
	if bindUnixSocket != "" {
		bindAddr = "unix:" + bindUnixSocket
	}
	cfg.Web = WebConfig{
		PublicURL:          webSection.Key("public_url").String(),
		BindAddr:           bindAddr,
		BindUnixSocket:     bindUnixSocket,
		CORSAllowedOrigins: splitCommaList(webSection.Key("cors_allowed_origins").String()),
		KillswitchChatAnnouncementsEnabled: parseBool01(webSection.Key("killswitch_chat_announcements_enabled").MustString("1")),
	}

	redisSection := file.Section("redis")
	cfg.Redis = RedisConfig{
		Addr:      redisSection.Key("addr").String(),
		Password:  redisSection.Key("password").String(),
		DB:        redisSection.Key("db").MustInt(0),
		KeyPrefix: redisSection.Key("key_prefix").MustString("dankbot"),
	}

	upDownSection := file.Section("updown")
	cfg.UpDown = UpDownConfig{
		Up:   upDownSection.Key("up").String(),
		Down: upDownSection.Key("down").String(),
	}

	twitchSection := file.Section("twitch")
	cfg.Twitch = TwitchConfig{
		ClientID:           twitchSection.Key("client_id").String(),
		ClientSecret:       twitchSection.Key("client_secret").String(),
		RedirectURI:        twitchSection.Key("redirect_uri").String(),
		ConnectRedirectURI: twitchSection.Key("connect_redirect_uri").String(),
		SendTransport:      strings.ToLower(strings.TrimSpace(twitchSection.Key("send_transport").MustString("irc"))),
	}

	twitchEventSubSection := file.Section("twitch_eventsub")

	syncInterval, err := parseDuration(twitchEventSubSection.Key("sync_interval").MustString("5m"))
	if err != nil {
		return nil, fmt.Errorf("parse twitch_eventsub.sync_interval: %w", err)
	}

	dedupeTTL, err := parseDuration(twitchEventSubSection.Key("dedupe_ttl").MustString("24h"))
	if err != nil {
		return nil, fmt.Errorf("parse twitch_eventsub.dedupe_ttl: %w", err)
	}

	cfg.TwitchEventSub = TwitchEventSubConfig{
		Enabled:      parseBool01(twitchEventSubSection.Key("enabled").String()),
		Transport:    strings.ToLower(strings.TrimSpace(twitchEventSubSection.Key("transport").MustString("webhook"))),
		Secret:       twitchEventSubSection.Key("secret").String(),
		CallbackURL:  twitchEventSubSection.Key("callback_url").String(),
		WebSocketURL: twitchEventSubSection.Key("websocket_url").String(),
		SyncInterval: syncInterval,
		DedupeTTL:    dedupeTTL,
	}

	openAISection := file.Section("openai")
	openAITimeout, err := parseDuration(openAISection.Key("timeout").MustString("5s"))
	if err != nil {
		return nil, fmt.Errorf("parse openai.timeout: %w", err)
	}

	cfg.OpenAI = OpenAIConfig{
		Enabled:           parseBool01(openAISection.Key("enabled").String()),
		APIKey:            openAISection.Key("api_key").String(),
		Model:             openAISection.Key("model").MustString("gpt-5-nano"),
		Timeout:           openAITimeout,
		KeywordValidation: parseBool01(openAISection.Key("keyword_validation").MustString("1")),
	}

	spotifySection := file.Section("spotify")
	cfg.Spotify = SpotifyConfig{
		Enabled:            parseBool01(spotifySection.Key("enabled").String()),
		ClientID:           spotifySection.Key("client_id").String(),
		ClientSecret:       spotifySection.Key("client_secret").String(),
		RedirectURI:        spotifySection.Key("redirect_uri").String(),
		LogSpotifyRequests: parseBool01(spotifySection.Key("log_spotify_requests").String()),
	}

	robloxSection := file.Section("roblox")
	cfg.Roblox = RobloxConfig{
		Enabled:      parseBool01(robloxSection.Key("enabled").String()),
		Cookie:       robloxSection.Key("cookie").String(),
		ClientID:     robloxSection.Key("client_id").String(),
		ClientSecret: robloxSection.Key("client_secret").String(),
		RedirectURI:  robloxSection.Key("redirect_uri").String(),
	}

	steamSection := file.Section("steam")
	cfg.Steam = SteamConfig{
		APIKey: steamSection.Key("api_key").String(),
		UserID: steamSection.Key("user_id").String(),
	}

	discordSection := file.Section("discord")
	cfg.Discord = DiscordConfig{
		Enabled:      parseBool01(discordSection.Key("enabled").String()),
		ClientID:     discordSection.Key("client_id").String(),
		ClientSecret: discordSection.Key("client_secret").String(),
		RedirectURI:  discordSection.Key("redirect_uri").String(),
		BotToken:     discordSection.Key("bot_token").String(),
		ModRoleID:    discordSection.Key("mod_role_id").String(),
	}

	streamlabsSection := file.Section("streamlabs")
	cfg.Streamlabs = StreamlabsConfig{
		Enabled:      parseBool01(streamlabsSection.Key("enabled").String()),
		ClientID:     streamlabsSection.Key("client_id").String(),
		ClientSecret: streamlabsSection.Key("client_secret").String(),
		RedirectURI:  streamlabsSection.Key("redirect_uri").String(),
	}

	streamElementsSection := file.Section("streamelements")
	cfg.StreamElements = StreamElementsConfig{
		Enabled:      parseBool01(streamElementsSection.Key("enabled").String()),
		ClientID:     streamElementsSection.Key("client_id").String(),
		ClientSecret: streamElementsSection.Key("client_secret").String(),
		RedirectURI:  streamElementsSection.Key("redirect_uri").String(),
	}

	rustLogSection := file.Section("rustlog")
	// Backward-compat: if [rustlog] is not configured, fall back to legacy [justlog].
	if len(rustLogSection.KeysHash()) == 0 {
		rustLogSection = file.Section("justlog")
	}
	cfg.RustLog = RustLogConfig{
		Enabled:    parseBool01(rustLogSection.Key("enabled").String()),
		BaseURL:    rustLogSection.Key("base_url").String(),
		APIKey:     rustLogSection.Key("api_key").String(),
		ConfigPath: rustLogSection.Key("config_path").String(),
	}

	workerSection := file.Section("worker")

	leaseTTL, err := parseDuration(workerSection.Key("lease_ttl").String())
	if err != nil {
		return nil, fmt.Errorf("parse worker.lease_ttl: %w", err)
	}

	heartbeatInterval, err := parseDuration(workerSection.Key("heartbeat_interval").String())
	if err != nil {
		return nil, fmt.Errorf("parse worker.heartbeat_interval: %w", err)
	}

	cfg.Worker = WorkerConfig{
		InstanceID:        workerSection.Key("instance_id").String(),
		LeaseTTL:          leaseTTL,
		HeartbeatInterval: heartbeatInterval,
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}
