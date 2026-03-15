package oauth

import (
	"bytes"
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
	defaultAuthorizeURL = "https://discord.com/oauth2/authorize"
	defaultTokenURL     = "https://discord.com/api/oauth2/token"
	defaultAPIBaseURL   = "https://discord.com/api"
	defaultPermissions  = "0"
)

type Client struct {
	httpClient   *http.Client
	clientID     string
	clientSecret string
	redirectURI  string
	authorizeURL string
	tokenURL     string
	apiBaseURL   string
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
		apiBaseURL:   defaultAPIBaseURL,
	}
}

func (c *Client) RedirectURI() string {
	return c.redirectURI
}

func (c *Client) AuthorizeURL(state string) (string, error) {
	if c.clientID == "" {
		return "", fmt.Errorf("discord client id is required")
	}
	if c.redirectURI == "" {
		return "", fmt.Errorf("discord redirect uri is required")
	}

	u, err := url.Parse(c.authorizeURL)
	if err != nil {
		return "", fmt.Errorf("parse discord authorize url: %w", err)
	}

	query := u.Query()
	query.Set("client_id", c.clientID)
	query.Set("redirect_uri", c.redirectURI)
	query.Set("response_type", "code")
	query.Set("scope", "identify bot applications.commands")
	query.Set("state", state)
	query.Set("permissions", defaultPermissions)
	u.RawQuery = query.Encode()

	return u.String(), nil
}

func (c *Client) ExchangeCode(ctx context.Context, code, redirectURI string) (*Token, error) {
	values := url.Values{}
	values.Set("client_id", c.clientID)
	values.Set("client_secret", c.clientSecret)
	values.Set("grant_type", "authorization_code")
	values.Set("code", strings.TrimSpace(code))
	values.Set("redirect_uri", strings.TrimSpace(redirectURI))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenURL, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create discord token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute discord token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return nil, fmt.Errorf("discord token request failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var token Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("decode discord token response: %w", err)
	}

	return &token, nil
}

func (c *Client) User(ctx context.Context, accessToken string) (*User, error) {
	endpoint := strings.TrimRight(c.apiBaseURL, "/") + "/users/@me"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create discord user request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(accessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute discord user request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return nil, fmt.Errorf("discord user request failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("decode discord user response: %w", err)
	}

	return &user, nil
}

func (c *Client) Revoke(ctx context.Context, accessToken string) error {
	values := url.Values{}
	values.Set("client_id", c.clientID)
	values.Set("client_secret", c.clientSecret)
	values.Set("token", strings.TrimSpace(accessToken))
	values.Set("token_type_hint", "access_token")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.apiBaseURL, "/")+"/oauth2/token/revoke", bytes.NewBufferString(values.Encode()))
	if err != nil {
		return fmt.Errorf("create discord revoke request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute discord revoke request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return fmt.Errorf("discord revoke request failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return nil
}
