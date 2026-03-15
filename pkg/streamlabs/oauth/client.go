package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultAuthorizeURL = "https://streamlabs.com/api/v2.0/authorize"
	defaultTokenURL     = "https://streamlabs.com/api/v2.0/token"
)

type Client struct {
	httpClient   *http.Client
	clientID     string
	clientSecret string
	redirectURI  string
	authorizeURL string
	tokenURL     string
}

func NewClient(httpClient *http.Client, clientID, clientSecret, redirectURI string) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}

	return &Client{
		httpClient:   httpClient,
		clientID:     strings.TrimSpace(clientID),
		clientSecret: strings.TrimSpace(clientSecret),
		redirectURI:  strings.TrimSpace(redirectURI),
		authorizeURL: defaultAuthorizeURL,
		tokenURL:     defaultTokenURL,
	}
}

func (c *Client) RedirectURI() string {
	return c.redirectURI
}

func (c *Client) AuthorizeURL(req AuthorizeRequest) (string, error) {
	redirectURI := strings.TrimSpace(req.RedirectURI)
	if redirectURI == "" {
		redirectURI = c.redirectURI
	}

	if c.clientID == "" {
		return "", fmt.Errorf("streamlabs client id is required")
	}
	if redirectURI == "" {
		return "", fmt.Errorf("streamlabs redirect uri is required")
	}

	u, err := url.Parse(c.authorizeURL)
	if err != nil {
		return "", fmt.Errorf("parse streamlabs authorize url: %w", err)
	}

	query := u.Query()
	query.Set("client_id", c.clientID)
	query.Set("redirect_uri", redirectURI)
	query.Set("response_type", "code")
	if len(req.Scopes) > 0 {
		query.Set("scope", strings.Join(req.Scopes, " "))
	}
	if state := strings.TrimSpace(req.State); state != "" {
		query.Set("state", state)
	}
	u.RawQuery = query.Encode()

	return u.String(), nil
}

func (c *Client) ExchangeCode(ctx context.Context, code, redirectURI string) (*Token, error) {
	if c.clientID == "" {
		return nil, fmt.Errorf("streamlabs client id is required")
	}
	if c.clientSecret == "" {
		return nil, fmt.Errorf("streamlabs client secret is required")
	}

	redirectURI = strings.TrimSpace(redirectURI)
	if redirectURI == "" {
		redirectURI = c.redirectURI
	}
	if redirectURI == "" {
		return nil, fmt.Errorf("streamlabs redirect uri is required")
	}

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", strings.TrimSpace(code))
	form.Set("redirect_uri", redirectURI)
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)

	return c.exchangeToken(ctx, form)
}

func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*Token, error) {
	if c.clientID == "" {
		return nil, fmt.Errorf("streamlabs client id is required")
	}
	if c.clientSecret == "" {
		return nil, fmt.Errorf("streamlabs client secret is required")
	}

	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, fmt.Errorf("streamlabs refresh token is required")
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)

	return c.exchangeToken(ctx, form)
}

func (c *Client) exchangeToken(ctx context.Context, form url.Values) (*Token, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create streamlabs token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute streamlabs token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return nil, fmt.Errorf("streamlabs token request failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var token Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("decode streamlabs token response: %w", err)
	}

	token.ReceivedAt = time.Now().UTC()
	return &token, nil
}
