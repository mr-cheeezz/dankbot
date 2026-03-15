package socket

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	socketio_client "github.com/zhouhui8915/go-socket.io-client"
)

type Client struct {
	url         string
	socketToken string
	client      *socketio_client.Client
	messages    chan Message
}

func NewClient(url, socketToken string) *Client {
	url = strings.TrimSpace(url)
	if url == "" {
		url = DefaultSocketURL
	}

	return &Client{
		url:         url,
		socketToken: strings.TrimSpace(socketToken),
		messages:    make(chan Message, 16),
	}
}

func (c *Client) Connect(ctx context.Context) error {
	_ = ctx

	if c.client != nil {
		return nil
	}

	opts := &socketio_client.Options{
		Transport: "websocket",
		Query: map[string]string{
			"token": c.socketToken,
		},
	}

	client, err := socketio_client.NewClient(c.url, opts)
	if err != nil {
		return fmt.Errorf("connect streamlabs socket: %w", err)
	}

	if err := client.On(EventName, func(payload any) {
		body, err := json.Marshal(payload)
		if err != nil {
			return
		}

		c.messages <- Message{
			Event: EventName,
			Data:  body,
		}
	}); err != nil {
		return fmt.Errorf("register streamlabs event handler: %w", err)
	}

	if err := client.On("disconnect", func() {
		close(c.messages)
	}); err != nil {
		return fmt.Errorf("register streamlabs disconnect handler: %w", err)
	}

	c.client = client
	return nil
}

func (c *Client) Messages() <-chan Message {
	return c.messages
}
