package redis

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type Client struct {
	Addr      string
	Password  string
	DB        int
	KeyPrefix string

	mu     sync.Mutex
	client *goredis.Client
}

var ErrKeyNotFound = errors.New("redis key not found")

type PubSubMessage struct {
	Channel string
	Payload string
}

type Subscription struct {
	pubsub   *goredis.PubSub
	messages chan PubSubMessage
}

func NewClient(addr, password string, db int, keyPrefix string) *Client {
	return &Client{
		Addr:      addr,
		Password:  password,
		DB:        db,
		KeyPrefix: keyPrefix,
	}
}

func (c *Client) Remember(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	client, err := c.rawClient()
	if err != nil {
		return false, err
	}

	added, err := client.SetNX(ctx, c.prefixed(key), "1", ttl).Result()
	if err != nil {
		return false, fmt.Errorf("remember redis key %q: %w", key, err)
	}

	return added, nil
}

func (c *Client) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	client, err := c.rawClient()
	if err != nil {
		return err
	}

	if err := client.Set(ctx, c.prefixed(key), value, ttl).Err(); err != nil {
		return fmt.Errorf("set redis key %q: %w", key, err)
	}

	return nil
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	client, err := c.rawClient()
	if err != nil {
		return "", err
	}

	value, err := client.Get(ctx, c.prefixed(key)).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return "", ErrKeyNotFound
		}
		return "", fmt.Errorf("get redis key %q: %w", key, err)
	}

	return value, nil
}

func (c *Client) Publish(ctx context.Context, channel, payload string) error {
	client, err := c.rawClient()
	if err != nil {
		return err
	}

	if err := client.Publish(ctx, c.prefixed(channel), payload).Err(); err != nil {
		return fmt.Errorf("publish redis channel %q: %w", channel, err)
	}

	return nil
}

func (c *Client) Subscribe(ctx context.Context, channel string) (*Subscription, error) {
	client, err := c.rawClient()
	if err != nil {
		return nil, err
	}

	pubsub := client.Subscribe(ctx, c.prefixed(channel))
	if _, err := pubsub.Receive(ctx); err != nil {
		_ = pubsub.Close()
		return nil, fmt.Errorf("subscribe redis channel %q: %w", channel, err)
	}

	subscription := &Subscription{
		pubsub:   pubsub,
		messages: make(chan PubSubMessage, 16),
	}

	go func() {
		defer close(subscription.messages)

		for message := range pubsub.Channel() {
			if message == nil {
				continue
			}

			subscription.messages <- PubSubMessage{
				Channel: message.Channel,
				Payload: message.Payload,
			}
		}
	}()

	return subscription, nil
}

func (c *Client) AcquireLock(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
	client, err := c.rawClient()
	if err != nil {
		return false, err
	}

	locked, err := client.SetNX(ctx, c.prefixed(key), value, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("acquire redis lock %q: %w", key, err)
	}

	return locked, nil
}

func (c *Client) ReleaseLock(ctx context.Context, key string, value string) error {
	client, err := c.rawClient()
	if err != nil {
		return err
	}

	const releaseScript = `
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
end
return 0
`

	if err := client.Eval(ctx, releaseScript, []string{c.prefixed(key)}, value).Err(); err != nil {
		return fmt.Errorf("release redis lock %q: %w", key, err)
	}

	return nil
}

func (c *Client) Delete(ctx context.Context, key string) error {
	client, err := c.rawClient()
	if err != nil {
		return err
	}

	if err := client.Del(ctx, c.prefixed(key)).Err(); err != nil {
		return fmt.Errorf("delete redis key %q: %w", key, err)
	}

	return nil
}

func (c *Client) Consume(ctx context.Context, key string) (string, error) {
	client, err := c.rawClient()
	if err != nil {
		return "", err
	}

	const consumeScript = `
local value = redis.call("get", KEYS[1])
if value then
	redis.call("del", KEYS[1])
	return value
end
return false
`

	result, err := client.Eval(ctx, consumeScript, []string{c.prefixed(key)}).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return "", ErrKeyNotFound
		}
		return "", fmt.Errorf("consume redis key %q: %w", key, err)
	}
	if result == nil {
		return "", ErrKeyNotFound
	}
	if value, ok := result.(bool); ok && !value {
		return "", ErrKeyNotFound
	}

	value, ok := result.(string)
	if !ok {
		return "", fmt.Errorf("consume redis key %q: unexpected result type %T", key, result)
	}

	return value, nil
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client == nil {
		return nil
	}

	err := c.client.Close()
	c.client = nil
	return err
}

func (s *Subscription) Messages() <-chan PubSubMessage {
	if s == nil {
		return nil
	}

	return s.messages
}

func (s *Subscription) Close() error {
	if s == nil || s.pubsub == nil {
		return nil
	}

	return s.pubsub.Close()
}

func (c *Client) rawClient() (*goredis.Client, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client != nil {
		return c.client, nil
	}

	client := goredis.NewClient(&goredis.Options{
		Addr:     c.Addr,
		Password: c.Password,
		DB:       c.DB,
	})

	c.client = client
	return c.client, nil
}

func (c *Client) prefixed(key string) string {
	key = strings.TrimSpace(key)
	if c.KeyPrefix == "" {
		return key
	}
	if key == "" {
		return c.KeyPrefix
	}

	return c.KeyPrefix + ":" + key
}
