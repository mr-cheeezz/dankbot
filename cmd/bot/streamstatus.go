package main

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/common/config"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	twitchhelix "github.com/mr-cheeezz/dankbot/pkg/twitch/helix"
	twitchoauth "github.com/mr-cheeezz/dankbot/pkg/twitch/oauth"
)

type streamStatusChecker struct {
	config   *config.Config
	accounts *postgres.TwitchAccountStore
	oauth    *twitchoauth.Service

	mu          sync.Mutex
	lastChecked time.Time
	cachedLive  bool
}

func newStreamStatusChecker(cfg *config.Config, accounts *postgres.TwitchAccountStore) *streamStatusChecker {
	return &streamStatusChecker{
		config:   cfg,
		accounts: accounts,
		oauth: twitchoauth.NewService(
			twitchoauth.NewClient(nil, cfg.Twitch.ClientID, cfg.Twitch.ClientSecret, strings.TrimSpace(cfg.Twitch.RedirectURI)),
			nil,
		),
	}
}

func (c *streamStatusChecker) IsLive(ctx context.Context) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now().UTC()
	if !c.lastChecked.IsZero() && now.Sub(c.lastChecked) < 30*time.Second {
		return c.cachedLive, nil
	}

	live, err := c.lookup(ctx)
	if err != nil {
		if !c.lastChecked.IsZero() {
			return c.cachedLive, nil
		}
		return false, err
	}

	c.lastChecked = now
	c.cachedLive = live
	return live, nil
}

func (c *streamStatusChecker) lookup(ctx context.Context) (bool, error) {
	streamerID := strings.TrimSpace(c.config.Main.StreamerID)
	if streamerID == "" {
		return false, nil
	}

	accessToken := ""
	if c.accounts != nil {
		account, err := c.accounts.Get(ctx, postgres.TwitchAccountKindBot)
		if err == nil && account != nil {
			accessToken = strings.TrimSpace(account.AccessToken)
		}
	}

	if accessToken == "" && c.oauth != nil {
		token, err := c.oauth.AppToken(ctx)
		if err != nil {
			return false, err
		}
		accessToken = strings.TrimSpace(token.AccessToken)
	}

	if accessToken == "" {
		return false, nil
	}

	client := twitchhelix.NewClient(c.config.Twitch.ClientID, accessToken)
	streams, err := client.GetStreamsByUserIDs(ctx, []string{streamerID})
	if err != nil {
		return false, err
	}

	return len(streams) > 0, nil
}
