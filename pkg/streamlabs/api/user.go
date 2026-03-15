package api

import (
	"context"
	"net/http"
)

func (c *Client) GetUser(ctx context.Context) (UserResponse, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/user", nil, nil)
	if err != nil {
		return nil, err
	}

	var response UserResponse
	if err := c.do(req, &response); err != nil {
		return nil, err
	}

	return response, nil
}
