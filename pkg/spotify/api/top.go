package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

func (c *Client) GetCurrentUserTopArtists(ctx context.Context, timeRange TimeRange, limit, offset int) (*TopArtistsPage, error) {
	query := url.Values{}
	if timeRange != "" {
		query.Set("time_range", string(timeRange))
	}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		query.Set("offset", strconv.Itoa(offset))
	}

	req, err := c.newRequest(ctx, http.MethodGet, "/me/top/artists", query)
	if err != nil {
		return nil, err
	}

	var page TopArtistsPage
	if err := c.do(req, &page); err != nil {
		return nil, err
	}

	return &page, nil
}

func (c *Client) GetCurrentUserTopTracks(ctx context.Context, timeRange TimeRange, limit, offset int) (*TopTracksPage, error) {
	query := url.Values{}
	if timeRange != "" {
		query.Set("time_range", string(timeRange))
	}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	if offset > 0 {
		query.Set("offset", strconv.Itoa(offset))
	}

	req, err := c.newRequest(ctx, http.MethodGet, "/me/top/tracks", query)
	if err != nil {
		return nil, err
	}

	var page TopTracksPage
	if err := c.do(req, &page); err != nil {
		return nil, err
	}

	return &page, nil
}

func (c *Client) GetCurrentUserTopItems(ctx context.Context, itemType TopType, timeRange TimeRange, limit, offset int) (any, error) {
	switch itemType {
	case TopTypeArtists:
		return c.GetCurrentUserTopArtists(ctx, timeRange, limit, offset)
	case TopTypeTracks:
		return c.GetCurrentUserTopTracks(ctx, timeRange, limit, offset)
	default:
		return nil, fmt.Errorf("unsupported spotify top item type %q", itemType)
	}
}
