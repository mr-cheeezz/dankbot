package oauth

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

const defaultStateTTL = 10 * time.Minute

type Service struct {
	client     *Client
	stateStore StateStore
	stateTTL   time.Duration
	now        func() time.Time

	siteLoginRedirectURI string
	connectRedirectURI   string

	mu           sync.Mutex
	cachedAppTok *Token
}

func NewService(client *Client, stateStore StateStore) *Service {
	if stateStore == nil {
		stateStore = NewMemoryStateStore()
	}

	return &Service{
		client:     client,
		stateStore: stateStore,
		stateTTL:   defaultStateTTL,
		now:        time.Now,
	}
}

func (s *Service) SiteLoginURL(ctx context.Context) (string, error) {
	return s.startFlow(ctx, FlowSiteLogin, SiteLoginScopes, false, SiteLoginClaims, false)
}

func (s *Service) StreamerConnectURL(ctx context.Context) (string, error) {
	return s.startFlow(ctx, FlowStreamerConnect, StreamerScopes, true, nil, false)
}

func (s *Service) BotConnectURL(ctx context.Context) (string, error) {
	return s.startFlow(ctx, FlowBotConnect, BotScopes, true, nil, false)
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

	validation, err := s.client.ValidateToken(ctx, token.AccessToken)
	if err != nil {
		return nil, err
	}

	result := &CallbackResult{
		Flow:        state.Flow,
		Token:       *token,
		Validation:  validation,
		RequestedAt: s.now().UTC(),
	}

	if state.Flow == FlowSiteLogin {
		// Only call the OIDC userinfo endpoint when openid was requested.
		// (We avoid this by default to keep consent screen minimal and not request email.)
		for _, scope := range state.Scopes {
			if strings.EqualFold(strings.TrimSpace(scope), "openid") {
				userInfo, err := s.client.UserInfo(ctx, token.AccessToken)
				if err != nil {
					return nil, err
				}
				result.UserInfo = userInfo
				break
			}
		}
	}

	return result, nil
}

func (s *Service) AppToken(ctx context.Context) (*Token, error) {
	s.mu.Lock()
	cached := s.cachedAppTok
	if cached != nil {
		expiresAt := cached.ExpiresAt()
		if expiresAt.IsZero() || time.Until(expiresAt) > time.Minute {
			tokenCopy := *cached
			s.mu.Unlock()
			return &tokenCopy, nil
		}
	}
	s.mu.Unlock()

	token, err := s.client.AppToken(ctx)
	if err != nil {
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.cachedAppTok != nil {
			expiresAt := s.cachedAppTok.ExpiresAt()
			if expiresAt.IsZero() || time.Until(expiresAt) > 0 {
				tokenCopy := *s.cachedAppTok
				return &tokenCopy, nil
			}
		}

		return nil, err
	}

	s.mu.Lock()
	s.cachedAppTok = token
	s.mu.Unlock()

	tokenCopy := *token
	return &tokenCopy, nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*Token, error) {
	return s.client.RefreshToken(ctx, refreshToken)
}

func (s *Service) ValidateToken(ctx context.Context, accessToken string) (*ValidateResponse, error) {
	return s.client.ValidateToken(ctx, accessToken)
}

func (s *Service) RevokeToken(ctx context.Context, accessToken string) error {
	return s.client.RevokeToken(ctx, accessToken)
}

func (s *Service) SetSiteLoginRedirectURI(redirectURI string) {
	s.siteLoginRedirectURI = strings.TrimSpace(redirectURI)
}

func (s *Service) SetConnectRedirectURI(redirectURI string) {
	s.connectRedirectURI = strings.TrimSpace(redirectURI)
}

func (s *Service) startFlow(ctx context.Context, flow Flow, scopes []string, forceVerify bool, claims *Claims, includeNonce bool) (string, error) {
	stateKey, err := randomToken(32)
	if err != nil {
		return "", fmt.Errorf("generate oauth state: %w", err)
	}

	authState := AuthorizationState{
		Flow:        flow,
		Scopes:      append([]string(nil), scopes...),
		RedirectURI: s.redirectURIForFlow(flow),
		ForceVerify: forceVerify,
		CreatedAt:   s.now().UTC(),
	}

	if includeNonce {
		nonce, err := randomToken(32)
		if err != nil {
			return "", fmt.Errorf("generate oauth nonce: %w", err)
		}
		authState.Nonce = nonce
	}

	if err := s.stateStore.Save(ctx, stateKey, authState, s.stateTTL); err != nil {
		return "", fmt.Errorf("store oauth state: %w", err)
	}

	return s.client.AuthorizeURL(AuthorizeRequest{
		RedirectURI: authState.RedirectURI,
		Scopes:      authState.Scopes,
		State:       stateKey,
		ForceVerify: authState.ForceVerify,
		Nonce:       authState.Nonce,
		Claims:      claims,
	})
}

func (s *Service) redirectURIForFlow(flow Flow) string {
	switch flow {
	case FlowSiteLogin:
		if value := strings.TrimSpace(s.siteLoginRedirectURI); value != "" {
			return value
		}
	default:
		if value := strings.TrimSpace(s.connectRedirectURI); value != "" {
			return value
		}
	}

	return s.client.RedirectURI()
}
