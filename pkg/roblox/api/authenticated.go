package api

import (
	"context"
	"net/http"
)

func (c *Client) GetAuthenticatedUser(ctx context.Context) (*AuthenticatedUser, error) {
	endpoint, err := buildURL(c.usersBaseURL, "/v1/users/authenticated", nil)
	if err != nil {
		return nil, err
	}

	req, err := c.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var user AuthenticatedUser
	if err := c.do(req, &user); err != nil {
		return nil, err
	}

	return &user, nil
}
