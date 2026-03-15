package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	redispkg "github.com/mr-cheeezz/dankbot/pkg/redis"
)

var ErrStateNotFound = errors.New("oauth state not found")

type StateStore interface {
	Save(ctx context.Context, key string, state AuthorizationState, ttl time.Duration) error
	Consume(ctx context.Context, key string) (*AuthorizationState, error)
}

type RedisStateStore struct {
	redis *redispkg.Client
}

func NewRedisStateStore(redis *redispkg.Client) *RedisStateStore {
	return &RedisStateStore{redis: redis}
}

func (s *RedisStateStore) Save(ctx context.Context, key string, state AuthorizationState, ttl time.Duration) error {
	if s.redis == nil {
		return fmt.Errorf("redis client is required")
	}

	payload, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal roblox oauth state: %w", err)
	}

	return s.redis.Set(ctx, stateStoreKey(key), string(payload), ttl)
}

func (s *RedisStateStore) Consume(ctx context.Context, key string) (*AuthorizationState, error) {
	if s.redis == nil {
		return nil, fmt.Errorf("redis client is required")
	}

	payload, err := s.redis.Consume(ctx, stateStoreKey(key))
	if err != nil {
		if errors.Is(err, redispkg.ErrKeyNotFound) {
			return nil, ErrStateNotFound
		}
		return nil, err
	}

	var state AuthorizationState
	if err := json.Unmarshal([]byte(payload), &state); err != nil {
		return nil, fmt.Errorf("unmarshal roblox oauth state: %w", err)
	}

	return &state, nil
}

func stateStoreKey(key string) string {
	return "roblox:oauth:state:" + key
}
