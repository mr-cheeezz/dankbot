package oauth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (c *Client) ExchangeCode(ctx context.Context, code, codeVerifier string) (*Token, error) {
	if c.clientID == "" {
		return nil, fmt.Errorf("roblox client id is required")
	}
	if c.clientSecret == "" {
		return nil, fmt.Errorf("roblox client secret is required")
	}
	if strings.TrimSpace(codeVerifier) == "" {
		return nil, fmt.Errorf("pkce code verifier is required")
	}

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)
	form.Set("code", strings.TrimSpace(code))
	form.Set("code_verifier", strings.TrimSpace(codeVerifier))

	req, err := c.newRequest(ctx, http.MethodPost, "/v1/token", nil, strings.NewReader(form.Encode()))
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
