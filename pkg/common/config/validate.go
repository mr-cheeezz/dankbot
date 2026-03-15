package config

import (
	"errors"
	"strings"
)

func (c *Config) Validate() error {
	var problems []string

	if strings.TrimSpace(c.Main.BotID) == "" {
		problems = append(problems, "main.bot_id is required")
	}
	if strings.TrimSpace(c.Main.StreamerID) == "" {
		problems = append(problems, "main.streamer_id is required")
	}
	if strings.TrimSpace(c.Main.AdminID) == "" {
		problems = append(problems, "main.admin_id is required")
	}
	if strings.TrimSpace(c.Main.DB) == "" {
		problems = append(problems, "main.db is required")
	}

	if strings.TrimSpace(c.Web.PublicURL) == "" {
		problems = append(problems, "web.public_url is required")
	}
	if strings.TrimSpace(c.Web.BindAddr) == "" {
		problems = append(problems, "web.bind_addr is required")
	}

	if strings.TrimSpace(c.Redis.Addr) == "" {
		problems = append(problems, "redis.addr is required")
	}
	if c.Redis.DB < 0 {
		problems = append(problems, "redis.db cannot be negative")
	}

	if strings.TrimSpace(c.Twitch.ClientID) == "" {
		problems = append(problems, "twitch.client_id is required")
	}
	if strings.TrimSpace(c.Twitch.ClientSecret) == "" {
		problems = append(problems, "twitch.client_secret is required")
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

	if c.Discord.Enabled {
		if strings.TrimSpace(c.Discord.ClientID) == "" {
			problems = append(problems, "discord.client_id is required when discord.enabled = 1")
		}
		if strings.TrimSpace(c.Discord.ClientSecret) == "" {
			problems = append(problems, "discord.client_secret is required when discord.enabled = 1")
		}
	}

	if strings.TrimSpace(c.Worker.InstanceID) == "" {
		problems = append(problems, "worker.instance_id is required")
	}
	if c.Worker.LeaseTTL <= 0 {
		problems = append(problems, "worker.lease_ttl must be greater than 0")
	}
	if c.Worker.HeartbeatInterval <= 0 {
		problems = append(problems, "worker.heartbeat_interval must be greater than 0")
	}
	if c.Worker.HeartbeatInterval >= c.Worker.LeaseTTL {
		problems = append(problems, "worker.heartbeat_interval must be less than worker.lease_ttl")
	}

	if len(problems) == 0 {
		return nil
	}

	return errors.New("config validation failed:\n - " + strings.Join(problems, "\n - "))
}
