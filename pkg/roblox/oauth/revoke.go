package oauth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func (c *Client) RevokeToken(ctx context.Context, refreshToken string) error {
	if c.clientID == "" {
		return fmt.Errorf("roblox client id is required")
	}
	if c.clientSecret == "" {
		return fmt.Errorf("roblox client secret is required")
	}

	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return fmt.Errorf("refresh token is required")
	}

	form := url.Values{}
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)
	form.Set("token", refreshToken)

	req, err := c.newRequest(ctx, http.MethodPost, "/v1/token/revoke", nil, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	return c.do(req, nil)
}
