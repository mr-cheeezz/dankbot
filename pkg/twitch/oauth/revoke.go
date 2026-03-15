package oauth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func (c *Client) RevokeToken(ctx context.Context, token string) error {
	if c.clientID == "" {
		return fmt.Errorf("twitch client id is required")
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return fmt.Errorf("token is required")
	}

	query := url.Values{}
	query.Set("client_id", c.clientID)
	query.Set("token", token)

	req, err := c.newRequest(ctx, http.MethodPost, "/revoke", query, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")

	return c.do(req, nil)
}
