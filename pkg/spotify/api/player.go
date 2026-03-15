package api

import (
	"context"
	"net/http"
	"net/url"
)

func (c *Client) GetCurrentlyPlaying(ctx context.Context, market string) (*CurrentlyPlaying, error) {
	query := url.Values{}
	if market != "" {
		query.Set("market", market)
	}

	req, err := c.newRequest(ctx, http.MethodGet, "/me/player/currently-playing", query)
	if err != nil {
		return nil, err
	}

	var playing CurrentlyPlaying
	if err := c.do(req, &playing); err != nil {
		return nil, err
	}

	return &playing, nil
}

func (c *Client) GetQueue(ctx context.Context) (*Queue, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/me/player/queue", nil)
	if err != nil {
		return nil, err
	}

	var queue Queue
	if err := c.do(req, &queue); err != nil {
		return nil, err
	}

	return &queue, nil
}
