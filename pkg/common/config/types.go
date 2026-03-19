package config

import "time"

type Config struct {
	Main           MainConfig
	Web            WebConfig
	Redis          RedisConfig
	UpDown         UpDownConfig
	Twitch         TwitchConfig
	TwitchEventSub TwitchEventSubConfig
	OpenAI         OpenAIConfig
	Spotify        SpotifyConfig
	Roblox         RobloxConfig
	Steam          SteamConfig
	Discord        DiscordConfig
	Streamlabs     StreamlabsConfig
	StreamElements StreamElementsConfig
	Worker         WorkerConfig
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
	BindUnixSocket     string
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
	ClientID           string
	ClientSecret       string
	RedirectURI        string
	ConnectRedirectURI string
	SendTransport      string
}

type TwitchEventSubConfig struct {
	Enabled      bool
	Transport    string
	Secret       string
	CallbackURL  string
	WebSocketURL string
	SyncInterval time.Duration
	DedupeTTL    time.Duration
}

type OpenAIConfig struct {
	Enabled           bool
	APIKey            string
	Model             string
	Timeout           time.Duration
	KeywordValidation bool
}

type SpotifyConfig struct {
	Enabled            bool
	ClientID           string
	ClientSecret       string
	RedirectURI        string
	LogSpotifyRequests bool
}

type RobloxConfig struct {
	Enabled      bool
	Cookie       string
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

type SteamConfig struct {
	APIKey string
	UserID string
}

type DiscordConfig struct {
	Enabled      bool
	ClientID     string
	ClientSecret string
	RedirectURI  string
	BotToken     string
	ModRoleID    string
}

type StreamlabsConfig struct {
	Enabled      bool
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

type StreamElementsConfig struct {
	Enabled      bool
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

type WorkerConfig struct {
	InstanceID        string
	LeaseTTL          time.Duration
	HeartbeatInterval time.Duration
}
