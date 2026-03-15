package api

import (
	"context"
	"net/http"
)

func (c *Client) GetSocketToken(ctx context.Context) (*SocketTokenResponse, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/socket/token", nil, nil)
	if err != nil {
		return nil, err
	}

	var response SocketTokenResponse
	if err := c.do(req, &response); err != nil {
		return nil, err
	}

	if response.SocketToken == "" {
		response.SocketToken = response.Token
	}

	return &response, nil
}
