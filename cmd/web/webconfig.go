package main

import (
	"fmt"
	"strings"

	"github.com/mr-cheeezz/dankbot/pkg/common/config"
)

func loadWebConfig(path string) (*config.Config, error) {
	cfg, err := config.Load(path)
	if err != nil {
		return nil, err
	}

	if err := validateWebConfig(cfg); err != nil {
		return nil, fmt.Errorf("validate web config: %w", err)
	}

	return cfg, nil
}

func validateWebConfig(cfg *config.Config) error {
	var problems []string

	if strings.TrimSpace(cfg.Main.DB) == "" {
		problems = append(problems, "main.db is required")
	}
	if strings.TrimSpace(cfg.Redis.Addr) == "" {
		problems = append(problems, "redis.addr is required")
	}
	if strings.TrimSpace(cfg.Web.PublicURL) == "" {
		problems = append(problems, "web.public_url is required")
	}
	if strings.TrimSpace(cfg.Web.BindAddr) == "" {
		problems = append(problems, "web.bind_addr is required")
	}
	if strings.TrimSpace(cfg.Twitch.ClientID) == "" {
		problems = append(problems, "twitch.client_id is required")
	}
	if strings.TrimSpace(cfg.Twitch.ClientSecret) == "" {
		problems = append(problems, "twitch.client_secret is required")
	}
	if cfg.TwitchEventSub.Enabled && strings.TrimSpace(cfg.TwitchEventSub.Secret) == "" {
		problems = append(problems, "twitch_eventsub.secret is required when twitch_eventsub.enabled = 1")
	}

	if len(problems) == 0 {
		return nil
	}

	return fmt.Errorf("%s", strings.Join(problems, "; "))
}
