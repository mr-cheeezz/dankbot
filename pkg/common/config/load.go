package config

import (
	"fmt"
	"strings"
	"time"

	ini "gopkg.in/ini.v1"
)

func phraseBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1":
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

func phraseDur(value string) (time.Duration, error) {
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
	cfg.Web = WebConfig{
		PublicURL:          webSection.Key("public_url").String(),
		BindAddr:           webSection.Key("bind_addr").String(),
		CORSAllowedOrigins: splitCommaList(webSection.Key("cors_allowed_origins").String()),
	}

	redisSection := file.Section("redis")
	cfg.Redis = RedisConfig{
		Addr:      redisSection.Key("addr").String(),
		Password:  redisSection.Key("password").String(),
		DB:        redisSection.Key("db").MustInt(0),
		KeyPrefix: redisSection.Key("Key_prefix").MustString("dankbot"),
	}

	upDownSection := file.Section("updown")
	cfg.UpDown = UpDownConfig{
		Up:   upDownSection.Key("up").String(),
		Down: upDownSection.Key("down").String(),
	}

	twitchSection := file.Section("twitch")
	cfg.Twitch = TwitchConfig{
		ClientID:     twitchSection.Key("client_id").String(),
		ClientSecret: twitchSection.Key("client_secret").String(),
		RedirectURI:  twitchSection.Key("redirect_uri").String(),
	}

	spotifySection := file.Section("spotify")
	cfg.Spotify = SpotifyConfig{
		Enabled:                 phraseBool(spotifySection.Key("enabled").String()),
		ClientID:                spotifySection.Key("client_id").String(),
		ClientSecret:            spotifySection.Key("client_secret").String(),
		RedirectURI:             spotifySection.Key("redirect_uri").String(),
		LogSpotifyRequests:      phraseBool(spotifySection.Key("log_spotify_requests").String()),
		SpotifyLoggingChannelID: spotifySection.Key("spotify_logging_channel_id").String(),
	}

	robloxSection := file.Section("roblox")
	cfg.Roblox = RobloxConfig{
		Enabled:      phraseBool(robloxSection.Key("enabled").String()),
		Cookie:       robloxSection.Key("cookie").String(),
		ClientID:     robloxSection.Key("client_id").String(),
		ClientSecret: robloxSection.Key("client_secret").String(),
		RedirectURI:  robloxSection.Key("redirect_uri").String(),
	}

	discordSection := file.Section("discord")
	cfg.Discord = DiscordConfig{
		Enabled:           phraseBool(discordSection.Key("enabled").String()),
		ClientID:          discordSection.Key("client_id").String(),
		ClientSecret:      discordSection.Key("client_secret").String(),
		RedirectURI:       discordSection.Key("redirect_uri").String(),
		BotToken:          discordSection.Key("bot_token").String(),
		LogsChannelID:     discordSection.Key("logs_channel_id").String(),
		RolePingChannelID: discordSection.Key("role_ping_channel_id").String(),
		ModRoleID:         discordSection.Key("mod_role_id").String(),
	}

	alertsSection := file.Section("alerts")
	cfg.Alerts = AlertsConfig{
		StreamlabsToken:     alertsSection.Key("streamlabs_token").String(),
		StreamlabsSocket:    alertsSection.Key("streamlabs_socket").String(),
		StreamElementsToken: alertsSection.Key("streamelements_token").String(),
	}

	workerSection := file.Section("worker")

	leaseTTL, err := phraseDur(workerSection.Key("lease_ttl").String())
	if err != nil {
		return nil, fmt.Errorf("parse worker.lease_ttl: %w", err)
	}

	heartbeatInterval, err := phraseDur(workerSection.Key("heartbeat_interval").String())
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
