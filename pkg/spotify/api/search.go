package api

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type trackSearchResponse struct {
	Tracks struct {
		Items []Track `json:"items"`
	} `json:"tracks"`
}

func (c *Client) SearchTracks(ctx context.Context, query string, limit int) ([]Track, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, nil
	}

	values := url.Values{}
	values.Set("q", query)
	values.Set("type", "track")
	if limit > 0 {
		values.Set("limit", strconv.Itoa(limit))
	}

	req, err := c.newRequest(ctx, http.MethodGet, "/search", values)
	if err != nil {
		return nil, err
	}

	var resp trackSearchResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}

	return resp.Tracks.Items, nil
}
