package bot

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
)

type Client struct {
	config   Config
	session  *discordgo.Session
	handlers Handlers
	errs     chan error
	once     sync.Once
}

func NewClient(cfg Config) (*Client, error) {
	token := strings.TrimSpace(cfg.BotToken)
	if token == "" {
		return nil, fmt.Errorf("discord bot token is required")
	}

	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("create discord session: %w", err)
	}

	session.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsDirectMessages |
		discordgo.IntentsMessageContent

	client := &Client{
		config:  cfg,
		session: session,
		errs:    make(chan error, 1),
	}
	client.registerHandlers()

	return client, nil
}

func (c *Client) RegisterHandlers(handlers Handlers) {
	c.handlers = handlers
}

func (c *Client) Start(ctx context.Context) error {
	if err := c.session.Open(); err != nil {
		return fmt.Errorf("open discord session: %w", err)
	}

	go func() {
		<-ctx.Done()
		_ = c.Stop(context.Background())
	}()

	return nil
}

func (c *Client) Stop(ctx context.Context) error {
	_ = ctx
	var err error
	c.once.Do(func() {
		err = c.session.Close()
		close(c.errs)
	})
	return err
}

func (c *Client) SendMessage(channelID, content string) error {
	channelID = strings.TrimSpace(channelID)
	content = strings.TrimSpace(content)
	if channelID == "" || content == "" {
		return nil
	}

	_, err := c.session.ChannelMessageSend(channelID, content)
	if err != nil {
		return fmt.Errorf("send discord message: %w", err)
	}

	return nil
}

func (c *Client) Errors() <-chan error {
	return c.errs
}
