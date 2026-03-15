package oauth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (c *Client) ExchangeCode(ctx context.Context, code, redirectURI string) (*Token, error) {
	if c.clientID == "" {
		return nil, fmt.Errorf("twitch client id is required")
	}
	if c.clientSecret == "" {
		return nil, fmt.Errorf("twitch client secret is required")
	}

	redirectURI = strings.TrimSpace(redirectURI)
	if redirectURI == "" {
		redirectURI = c.redirectURI
	}
	if redirectURI == "" {
		return nil, fmt.Errorf("twitch redirect uri is required")
	}

	form := url.Values{}
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)
	form.Set("code", strings.TrimSpace(code))
	form.Set("grant_type", "authorization_code")
	form.Set("redirect_uri", redirectURI)

	return c.exchangeToken(ctx, form)
}

func (c *Client) exchangeToken(ctx context.Context, form url.Values) (*Token, error) {
	req, err := c.newRequest(ctx, http.MethodPost, "/token", nil, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	var token Token
	if err := c.do(req, &token); err != nil {
		return nil, err
	}

	token.ReceivedAt = time.Now().UTC()
	return &token, nil
}
