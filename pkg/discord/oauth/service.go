package oauth

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const defaultStateTTL = 10 * time.Minute

type Service struct {
	client     *Client
	stateStore StateStore
	stateTTL   time.Duration
	now        func() time.Time
}

func NewService(client *Client, stateStore StateStore) *Service {
	return &Service{
		client:     client,
		stateStore: stateStore,
		stateTTL:   defaultStateTTL,
		now:        time.Now,
	}
}

func (s *Service) JoinURL(ctx context.Context) (string, error) {
	stateKey, err := randomToken(32)
	if err != nil {
		return "", fmt.Errorf("generate discord oauth state: %w", err)
	}

	state := AuthorizationState{
		RedirectURI: s.client.RedirectURI(),
		CreatedAt:   s.now().UTC(),
	}

	if err := s.stateStore.Save(ctx, stateKey, state, s.stateTTL); err != nil {
		return "", fmt.Errorf("store discord oauth state: %w", err)
	}

	return s.client.AuthorizeURL(stateKey)
}

func (s *Service) HandleCallback(ctx context.Context, stateKey, code, guildID, permissions string) (*CallbackResult, error) {
	state, err := s.stateStore.Consume(ctx, stateKey)
	if err != nil {
		return nil, err
	}

	token, err := s.client.ExchangeCode(ctx, code, state.RedirectURI)
	if err != nil {
		return nil, err
	}

	user, err := s.client.User(ctx, token.AccessToken)
	if err != nil {
		return nil, err
	}

	return &CallbackResult{
		Token:       *token,
		User:        user,
		GuildID:     strings.TrimSpace(guildID),
		Permissions: strings.TrimSpace(permissions),
		RequestedAt: s.now().UTC(),
	}, nil
}
