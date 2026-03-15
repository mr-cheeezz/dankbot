package api

import (
	"context"
	"net/http"
)

type ChannelMeResponse struct {
	ID          string `json:"_id"`
	Provider    string `json:"provider"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Alias       string `json:"alias"`
}

type UserCurrentResponse struct {
	ID          string `json:"_id"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
}

func (c *Client) GetChannelMe(ctx context.Context) (*ChannelMeResponse, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/channels/me")
	if err != nil {
		return nil, err
	}

	var response ChannelMeResponse
	if err := c.do(req, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *Client) GetCurrentUser(ctx context.Context) (*UserCurrentResponse, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/users/current")
	if err != nil {
		return nil, err
	}

	var response UserCurrentResponse
	if err := c.do(req, &response); err != nil {
		return nil, err
	}

	return &response, nil
}
