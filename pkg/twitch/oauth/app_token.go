package oauth

import (
	"context"
	"fmt"
	"net/url"
)

func (c *Client) AppToken(ctx context.Context) (*Token, error) {
	if c.clientID == "" {
		return nil, fmt.Errorf("twitch client id is required")
	}
	if c.clientSecret == "" {
		return nil, fmt.Errorf("twitch client secret is required")
	}

	form := url.Values{}
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)
	form.Set("grant_type", "client_credentials")

	return c.exchangeToken(ctx, form)
}
