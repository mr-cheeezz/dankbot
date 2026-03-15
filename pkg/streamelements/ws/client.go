package ws

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"nhooyr.io/websocket"
)

type Client struct {
	gatewayURL string
	token      string
	tokenType  string

	mu   sync.Mutex
	conn *websocket.Conn
}

func NewClient(gatewayURL, token, tokenType string) *Client {
	gatewayURL = strings.TrimSpace(gatewayURL)
	if gatewayURL == "" {
		gatewayURL = DefaultGatewayURL
	}

	return &Client{
		gatewayURL: gatewayURL,
		token:      strings.TrimSpace(token),
		tokenType:  strings.TrimSpace(tokenType),
	}
}

func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return nil
	}

	conn, _, err := websocket.Dial(ctx, c.gatewayURL, nil)
	if err != nil {
		return fmt.Errorf("connect streamelements websocket: %w", err)
	}

	c.conn = conn
	return nil
}

func (c *Client) SubscribeChannelActivities(ctx context.Context, room string) error {
	return c.Subscribe(ctx, TopicChannelActivities, room)
}

func (c *Client) SubscribeChannelTips(ctx context.Context, room string) error {
	return c.Subscribe(ctx, TopicChannelTips, room)
}

func (c *Client) SubscribeChannelTipModeration(ctx context.Context, room string) error {
	return c.Subscribe(ctx, TopicChannelTipMods, room)
}

func (c *Client) SubscribeChannelStreamStatus(ctx context.Context, room string) error {
	return c.Subscribe(ctx, TopicChannelStream, room)
}

func (c *Client) Subscribe(ctx context.Context, topic, room string) error {
	if err := c.Connect(ctx); err != nil {
		return err
	}

	nonce, err := randomNonce(24)
	if err != nil {
		return fmt.Errorf("generate streamelements nonce: %w", err)
	}

	request := SubscribeRequest{
		Type:  "subscribe",
		Nonce: nonce,
		Data: SubscribeRequestData{
			Topic:     strings.TrimSpace(topic),
			Room:      strings.TrimSpace(room),
			Token:     c.token,
			TokenType: c.tokenType,
		},
	}

	return c.WriteJSON(ctx, request)
}

func (c *Client) WriteJSON(ctx context.Context, payload any) error {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return fmt.Errorf("streamelements websocket is not connected")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal streamelements websocket payload: %w", err)
	}

	if err := conn.Write(ctx, websocket.MessageText, body); err != nil {
		return fmt.Errorf("write streamelements websocket message: %w", err)
	}

	return nil
}

func (c *Client) Read(ctx context.Context) (*Message, error) {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return nil, fmt.Errorf("streamelements websocket is not connected")
	}

	_, body, err := conn.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("read streamelements websocket message: %w", err)
	}

	var message Message
	if err := json.Unmarshal(body, &message); err != nil {
		return nil, fmt.Errorf("decode streamelements websocket message: %w", err)
	}

	return &message, nil
}

func (c *Client) Close(status websocket.StatusCode, reason string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil
	}

	err := c.conn.Close(status, reason)
	c.conn = nil
	return err
}

func randomNonce(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}
