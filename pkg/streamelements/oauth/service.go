package oauth

import (
	"context"
	"fmt"
	"time"
)

const defaultStateTTL = 10 * time.Minute

type Flow string

const (
	FlowStreamerConnect Flow = "streamer_connect"
)

type AuthorizationState struct {
	Flow        Flow      `json:"flow"`
	Scopes      []string  `json:"scopes"`
	RedirectURI string    `json:"redirect_uri"`
	CreatedAt   time.Time `json:"created_at"`
}

type CallbackResult struct {
	Flow        Flow      `json:"flow"`
	Token       Token     `json:"token"`
	RequestedAt time.Time `json:"requested_at"`
}

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

func (s *Service) StreamerConnectURL(ctx context.Context) (string, error) {
	stateKey, err := randomToken(32)
	if err != nil {
		return "", fmt.Errorf("generate streamelements oauth state: %w", err)
	}

	authState := AuthorizationState{
		Flow:        FlowStreamerConnect,
		RedirectURI: s.client.RedirectURI(),
		CreatedAt:   s.now().UTC(),
	}

	if err := s.stateStore.Save(ctx, stateKey, authState, s.stateTTL); err != nil {
		return "", fmt.Errorf("store streamelements oauth state: %w", err)
	}

	return s.client.AuthorizeURL(AuthorizeRequest{
		RedirectURI: authState.RedirectURI,
		Scopes:      authState.Scopes,
		State:       stateKey,
	})
}

func (s *Service) HandleCallback(ctx context.Context, stateKey, code string) (*CallbackResult, error) {
	state, err := s.stateStore.Consume(ctx, stateKey)
	if err != nil {
		return nil, err
	}

	token, err := s.client.ExchangeCode(ctx, code, state.RedirectURI)
	if err != nil {
		return nil, err
	}

	return &CallbackResult{
		Flow:        state.Flow,
		Token:       *token,
		RequestedAt: s.now().UTC(),
	}, nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*Token, error) {
	return s.client.RefreshToken(ctx, refreshToken)
}
