package api

import (
	"context"
	"net/http"
	"net/url"
	"strings"
)

func (c *Client) GetPoints(ctx context.Context, username, channel string) (PointsResponse, error) {
	query := url.Values{}
	if value := strings.TrimSpace(username); value != "" {
		query.Set("username", value)
	}
	if value := strings.TrimSpace(channel); value != "" {
		query.Set("channel", value)
	}

	req, err := c.newRequest(ctx, http.MethodGet, "/points", query, nil)
	if err != nil {
		return nil, err
	}

	var response PointsResponse
	if err := c.do(req, &response); err != nil {
		return nil, err
	}

	return response, nil
}
