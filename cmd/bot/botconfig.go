package main

import (
	"fmt"
	"strings"

	"github.com/mr-cheeezz/dankbot/pkg/common/config"
)

func loadBotConfig(path string) (*config.Config, error) {
	cfg, err := config.Load(path)
	if err != nil {
		return nil, err
	}

	if err := validateBotConfig(cfg); err != nil {
		return nil, fmt.Errorf("validate bot config: %w", err)
	}

	return cfg, nil
}

func validateBotConfig(cfg *config.Config) error {
	var problems []string

	if strings.TrimSpace(cfg.Main.BotID) == "" {
		problems = append(problems, "main.bot_id is required")
	}
	if strings.TrimSpace(cfg.Main.StreamerID) == "" {
		problems = append(problems, "main.streamer_id is required")
	}
	if strings.TrimSpace(cfg.Main.AdminID) == "" {
		problems = append(problems, "main.admin_id is required")
	}
	if strings.TrimSpace(cfg.Main.DB) == "" {
		problems = append(problems, "main.db is required")
	}
	if strings.TrimSpace(cfg.Twitch.ClientID) == "" {
		problems = append(problems, "twitch.client_id is required")
	}
	if strings.TrimSpace(cfg.Twitch.ClientSecret) == "" {
		problems = append(problems, "twitch.client_secret is required")
	}
	if transport := strings.TrimSpace(cfg.Twitch.SendTransport); transport == "" {
		problems = append(problems, "twitch.send_transport is required")
	}
	if len(problems) == 0 {
		return nil
	}

	return fmt.Errorf("%s", strings.Join(problems, "; "))
}
