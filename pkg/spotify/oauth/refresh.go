package oauth

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*Token, error) {
	if c.clientID == "" {
		return nil, fmt.Errorf("spotify client id is required")
	}
	if c.clientSecret == "" {
		return nil, fmt.Errorf("spotify client secret is required")
	}

	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, fmt.Errorf("spotify refresh token is required")
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)

	return c.exchangeToken(ctx, form)
}
