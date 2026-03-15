package config

import "time"

type Config struct {
	Main    MainConfig
	Web     WebConfig
	Redis   RedisConfig
	UpDown  UpDownConfig
	Twitch  TwitchConfig
	Spotify SpotifyConfig
	Roblox  RobloxConfig
	Discord DiscordConfig
	Alerts  AlertsConfig
	Worker  WorkerConfig
}

type MainConfig struct {
	BotID      string
	StreamerID string
	AdminID    string
	DB         string
}

type WebConfig struct {
	PublicURL          string
	BindAddr           string
	CORSAllowedOrigins []string
}

type RedisConfig struct {
	Addr      string
	Password  string
	DB        int
	KeyPrefix string
}

type UpDownConfig struct {
	Up   string
	Down string
}

type TwitchConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

type SpotifyConfig struct {
	Enabled                 bool
	ClientID                string
	ClientSecret            string
	RedirectURI             string
	LogSpotifyRequests      bool
	SpotifyLoggingChannelID string
}

type RobloxConfig struct {
	Enabled      bool
	Cookie       string
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

type DiscordConfig struct {
	Enabled           bool
	ClientID          string
	ClientSecret      string
	RedirectURI       string
	BotToken          string
	LogsChannelID     string
	RolePingChannelID string
	ModRoleID         string
}

type AlertsConfig struct {
	StreamlabsToken     string
	StreamlabsSocket    string
	StreamElementsToken string
}

type WorkerConfig struct {
	InstanceID        string
	LeaseTTL          time.Duration
	HeartbeatInterval time.Duration
}
