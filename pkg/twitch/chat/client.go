package chat

import (
	"context"

	twitch "github.com/gempir/go-twitch-irc/v4"
)

type Config struct {
	BotLogin   string
	OAuthToken string
	Channels   []string
}

type Client struct {
	config   Config
	irc      *twitch.Client
	handlers Handlers
	sync     ChannelSync
	errs     chan error
}

func NewClient(cfg Config) *Client {
	ircClient := twitch.NewClient(cfg.BotLogin, cfg.OAuthToken)

	client := &Client{
		config: cfg,
		irc:    ircClient,
		sync:   NewChannelSync(),
		errs:   make(chan error, 1),
	}

	client.sync.SetOwnedChannels(cfg.Channels)
	client.registerIRCHandlers()

	return client
}

func (c *Client) Start(ctx context.Context) error {
	for _, channel := range c.sync.OwnedChannels() {
		c.irc.Join(channel)
	}

	go func() {
		c.errs <- c.irc.Connect()
	}()

	return nil
}

func (c *Client) Stop(ctx context.Context) error {
	_ = ctx
	return c.irc.Disconnect()
}

func (c *Client) Errors() <-chan error {
	return c.errs
}

func (c *Client) Config() Config {
	return c.config
}

func (c *Client) SetOAuthToken(token string) {
	c.config.OAuthToken = token
	c.irc.SetIRCToken(token)
}
