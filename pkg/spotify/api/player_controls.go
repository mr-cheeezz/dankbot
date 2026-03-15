package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (c *Client) GetRecentlyPlayed(ctx context.Context, limit int, after, before string) (*RecentlyPlayedPage, error) {
	query := url.Values{}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	if strings.TrimSpace(after) != "" {
		query.Set("after", strings.TrimSpace(after))
	}
	if strings.TrimSpace(before) != "" {
		query.Set("before", strings.TrimSpace(before))
	}

	req, err := c.newRequest(ctx, http.MethodGet, "/me/player/recently-played", query)
	if err != nil {
		return nil, err
	}

	var page RecentlyPlayedPage
	if err := c.do(req, &page); err != nil {
		return nil, err
	}

	return &page, nil
}

func (c *Client) AddToQueue(ctx context.Context, itemURI, deviceID string) error {
	itemURI = strings.TrimSpace(itemURI)
	if itemURI == "" {
		return fmt.Errorf("spotify item uri is required")
	}

	query := url.Values{}
	query.Set("uri", itemURI)
	if strings.TrimSpace(deviceID) != "" {
		query.Set("device_id", strings.TrimSpace(deviceID))
	}

	req, err := c.newRequest(ctx, http.MethodPost, "/me/player/queue", query)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) PausePlayback(ctx context.Context, deviceID string) error {
	query := url.Values{}
	if strings.TrimSpace(deviceID) != "" {
		query.Set("device_id", strings.TrimSpace(deviceID))
	}

	req, err := c.newRequest(ctx, http.MethodPut, "/me/player/pause", query)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) ResumePlayback(ctx context.Context, deviceID string) error {
	query := url.Values{}
	if strings.TrimSpace(deviceID) != "" {
		query.Set("device_id", strings.TrimSpace(deviceID))
	}

	req, err := c.newRequest(ctx, http.MethodPut, "/me/player/play", query)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) SkipNext(ctx context.Context, deviceID string) error {
	query := url.Values{}
	if strings.TrimSpace(deviceID) != "" {
		query.Set("device_id", strings.TrimSpace(deviceID))
	}

	req, err := c.newRequest(ctx, http.MethodPost, "/me/player/next", query)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) SkipPrevious(ctx context.Context, deviceID string) error {
	query := url.Values{}
	if strings.TrimSpace(deviceID) != "" {
		query.Set("device_id", strings.TrimSpace(deviceID))
	}

	req, err := c.newRequest(ctx, http.MethodPost, "/me/player/previous", query)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) SetVolume(ctx context.Context, volumePercent int, deviceID string) error {
	if volumePercent < 0 || volumePercent > 100 {
		return fmt.Errorf("spotify volume percent must be between 0 and 100")
	}

	query := url.Values{}
	query.Set("volume_percent", strconv.Itoa(volumePercent))
	if strings.TrimSpace(deviceID) != "" {
		query.Set("device_id", strings.TrimSpace(deviceID))
	}

	req, err := c.newRequest(ctx, http.MethodPut, "/me/player/volume", query)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}
