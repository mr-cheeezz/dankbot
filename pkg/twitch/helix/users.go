package helix

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func (c *Client) GetUsersByIDs(ctx context.Context, ids []string) ([]User, error) {
	query := url.Values{}

	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id != "" {
			query.Add("id", id)
		}
	}

	if len(query["id"]) == 0 {
		return nil, fmt.Errorf("at least one user id is required")
	}

	req, err := c.newRequest(ctx, http.MethodGet, "/users", query)
	if err != nil {
		return nil, err
	}

	var resp usersResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}

	return resp.Data, nil
}

func (c *Client) GetUsersByLogins(ctx context.Context, logins []string) ([]User, error) {
	query := url.Values{}

	for _, login := range logins {
		login = strings.TrimSpace(login)
		if login != "" {
			query.Add("login", login)
		}
	}

	if len(query["login"]) == 0 {
		return nil, fmt.Errorf("at least one user login is required")
	}

	req, err := c.newRequest(ctx, http.MethodGet, "/users", query)
	if err != nil {
		return nil, err
	}

	var resp usersResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}

	return resp.Data, nil
}
