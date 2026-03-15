package oauth

import (
	"context"
	"fmt"
	"time"
)

const defaultStateTTL = 10 * time.Minute

var StreamerScopes = []string{"openid", "profile"}

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
		return "", fmt.Errorf("generate oauth state: %w", err)
	}

	verifier, err := newCodeVerifier()
	if err != nil {
		return "", fmt.Errorf("generate pkce verifier: %w", err)
	}

	authState := AuthorizationState{
		Flow:         FlowStreamerConnect,
		Scopes:       append([]string(nil), StreamerScopes...),
		RedirectURI:  s.client.RedirectURI(),
		CodeVerifier: verifier,
		CreatedAt:    s.now().UTC(),
	}

	if err := s.stateStore.Save(ctx, stateKey, authState, s.stateTTL); err != nil {
		return "", fmt.Errorf("store oauth state: %w", err)
	}

	return s.client.AuthorizeURL(AuthorizeRequest{
		RedirectURI:   authState.RedirectURI,
		Scopes:        authState.Scopes,
		State:         stateKey,
		CodeChallenge: codeChallenge(verifier),
	})
}

func (s *Service) HandleCallback(ctx context.Context, stateKey, code string) (*CallbackResult, error) {
	state, err := s.stateStore.Consume(ctx, stateKey)
	if err != nil {
		return nil, err
	}

	token, err := s.client.ExchangeCode(ctx, code, state.CodeVerifier)
	if err != nil {
		return nil, err
	}

	userInfo, err := s.client.UserInfo(ctx, token.AccessToken)
	if err != nil {
		return nil, err
	}

	return &CallbackResult{
		Flow:        state.Flow,
		Token:       *token,
		UserInfo:    userInfo,
		RequestedAt: s.now().UTC(),
	}, nil
}

func (s *Service) RevokeToken(ctx context.Context, refreshToken string) error {
	return s.client.RevokeToken(ctx, refreshToken)
}

func randomToken(size int) (string, error) {
	return newCodeVerifierWithSize(size)
}
