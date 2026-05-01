package config

import (
	"errors"
	"strings"
)

func (c *Config) Validate() error {
	var problems []string

	if c.Redis.DB < 0 {
		problems = append(problems, "redis.db cannot be negative")
	}

	if c.TwitchEventSub.Enabled {
		transport := strings.TrimSpace(c.TwitchEventSub.Transport)
		if transport == "" {
			transport = "webhook"
		}
		if transport != "webhook" && transport != "websocket" {
			problems = append(problems, "twitch_eventsub.transport must be either webhook or websocket")
		}
		if transport == "webhook" && strings.TrimSpace(c.TwitchEventSub.Secret) == "" {
			problems = append(problems, "twitch_eventsub.secret is required when twitch_eventsub.enabled = 1 and transport = webhook")
		}
		if c.TwitchEventSub.SyncInterval <= 0 {
			problems = append(problems, "twitch_eventsub.sync_interval must be greater than 0")
		}
		if c.TwitchEventSub.DedupeTTL <= 0 {
			problems = append(problems, "twitch_eventsub.dedupe_ttl must be greater than 0")
		}
	}

	if c.OpenAI.Enabled {
		if strings.TrimSpace(c.OpenAI.APIKey) == "" {
			problems = append(problems, "openai.api_key is required when openai.enabled = 1")
		}
		if strings.TrimSpace(c.OpenAI.Model) == "" {
			problems = append(problems, "openai.model is required when openai.enabled = 1")
		}
		if c.OpenAI.Timeout <= 0 {
			problems = append(problems, "openai.timeout must be greater than 0")
		}
	}

	if transport := strings.TrimSpace(c.Twitch.SendTransport); transport != "" && transport != "irc" && transport != "helix" {
		problems = append(problems, "twitch.send_transport must be either irc or helix")
	}

	if c.Spotify.Enabled {
		if strings.TrimSpace(c.Spotify.ClientID) == "" {
			problems = append(problems, "spotify.client_id is required when spotify.enabled = 1")
		}
		if strings.TrimSpace(c.Spotify.ClientSecret) == "" {
			problems = append(problems, "spotify.client_secret is required when spotify.enabled = 1")
		}
	}

	if c.Roblox.Enabled {
		if strings.TrimSpace(c.Roblox.ClientID) == "" {
			problems = append(problems, "roblox.client_id is required when roblox.enabled = 1")
		}
		if strings.TrimSpace(c.Roblox.ClientSecret) == "" {
			problems = append(problems, "roblox.client_secret is required when roblox.enabled = 1")
		}
	}

	steamAPIKey := strings.TrimSpace(c.Steam.APIKey)
	steamUserID := strings.TrimSpace(c.Steam.UserID)
	if steamAPIKey != "" && steamUserID == "" {
		problems = append(problems, "steam.user_id is required when steam.api_key is set")
	}
	if steamUserID != "" && steamAPIKey == "" {
		problems = append(problems, "steam.api_key is required when steam.user_id is set")
	}

	if c.Discord.Enabled {
		if strings.TrimSpace(c.Discord.ClientID) == "" {
			problems = append(problems, "discord.client_id is required when discord.enabled = 1")
		}
		if strings.TrimSpace(c.Discord.ClientSecret) == "" {
			problems = append(problems, "discord.client_secret is required when discord.enabled = 1")
		}
	}

	if c.Streamlabs.Enabled {
		if strings.TrimSpace(c.Streamlabs.ClientID) == "" {
			problems = append(problems, "streamlabs.client_id is required when streamlabs.enabled = 1")
		}
		if strings.TrimSpace(c.Streamlabs.ClientSecret) == "" {
			problems = append(problems, "streamlabs.client_secret is required when streamlabs.enabled = 1")
		}
	}

	if c.StreamElements.Enabled {
		if strings.TrimSpace(c.StreamElements.ClientID) == "" {
			problems = append(problems, "streamelements.client_id is required when streamelements.enabled = 1")
		}
		if strings.TrimSpace(c.StreamElements.ClientSecret) == "" {
			problems = append(problems, "streamelements.client_secret is required when streamelements.enabled = 1")
		}
	}

	if c.RustLog.Enabled {
		if strings.TrimSpace(c.RustLog.BaseURL) == "" && strings.TrimSpace(c.RustLog.ConfigPath) == "" {
			problems = append(problems, "rustlog.base_url or rustlog.config_path is required when rustlog.enabled = 1")
		}
	}

	if c.Worker.HeartbeatInterval > 0 && c.Worker.LeaseTTL > 0 && c.Worker.HeartbeatInterval >= c.Worker.LeaseTTL {
		problems = append(problems, "worker.heartbeat_interval must be less than worker.lease_ttl")
	}

	if len(problems) == 0 {
		return nil
	}

	return errors.New("config validation failed:\n - " + strings.Join(problems, "\n - "))
}
