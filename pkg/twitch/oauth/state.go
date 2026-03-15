package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

var (
	ErrStateNotFound = errors.New("oauth state not found")
	ErrStateExpired  = errors.New("oauth state expired")
)

type StateStore interface {
	Save(ctx context.Context, key string, state AuthorizationState, ttl time.Duration) error
	Consume(ctx context.Context, key string) (*AuthorizationState, error)
}

type memoryStateEntry struct {
	state     AuthorizationState
	expiresAt time.Time
}

type MemoryStateStore struct {
	mu      sync.Mutex
	entries map[string]memoryStateEntry
}

func NewMemoryStateStore() *MemoryStateStore {
	return &MemoryStateStore{
		entries: map[string]memoryStateEntry{},
	}
}

func (s *MemoryStateStore) Save(ctx context.Context, key string, state AuthorizationState, ttl time.Duration) error {
	_ = ctx

	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries[key] = memoryStateEntry{
		state:     state,
		expiresAt: time.Now().UTC().Add(ttl),
	}

	return nil
}

func (s *MemoryStateStore) Consume(ctx context.Context, key string) (*AuthorizationState, error) {
	_ = ctx

	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.entries[key]
	if !ok {
		return nil, ErrStateNotFound
	}

	delete(s.entries, key)

	if time.Now().UTC().After(entry.expiresAt) {
		return nil, ErrStateExpired
	}

	state := entry.state
	return &state, nil
}

func randomToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}
