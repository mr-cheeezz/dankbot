package api

import (
	"context"
	"net/http"
)

func (c *Client) GetCurrentUserProfile(ctx context.Context) (*UserProfile, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/me", nil)
	if err != nil {
		return nil, err
	}

	var profile UserProfile
	if err := c.do(req, &profile); err != nil {
		return nil, err
	}

	return &profile, nil
}
