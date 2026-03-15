package api

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

func (c *Client) GetCurrentUserPlaylists(ctx context.Context, limit, offset int) (*PlaylistsPage, error) {
	query := url.Values{}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		query.Set("offset", strconv.Itoa(offset))
	}

	req, err := c.newRequest(ctx, http.MethodGet, "/me/playlists", query)
	if err != nil {
		return nil, err
	}

	var page PlaylistsPage
	if err := c.do(req, &page); err != nil {
		return nil, err
	}

	return &page, nil
}
