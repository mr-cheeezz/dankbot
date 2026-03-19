package helix

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ChannelSearchResult struct {
	ID               string    `json:"id"`
	BroadcasterLogin string    `json:"broadcaster_login"`
	DisplayName      string    `json:"display_name"`
	GameID           string    `json:"game_id"`
	GameName         string    `json:"game_name"`
	IsLive           bool      `json:"is_live"`
	ThumbnailURL     string    `json:"thumbnail_url"`
	Title            string    `json:"title"`
	StartedAt        time.Time `json:"started_at"`
}

type searchChannelsResponse struct {
	Data []ChannelSearchResult `json:"data"`
}

type CategorySearchResult struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	BoxArtURL string `json:"box_art_url"`
}

type searchCategoriesResponse struct {
	Data []CategorySearchResult `json:"data"`
}

func (c *Client) SearchChannels(ctx context.Context, queryText string, first int, liveOnly bool) ([]ChannelSearchResult, error) {
	queryText = strings.TrimSpace(queryText)
	if queryText == "" {
		return nil, fmt.Errorf("search query is required")
	}

	query := url.Values{}
	query.Set("query", queryText)
	if first > 0 {
		query.Set("first", strconv.Itoa(first))
	}
	if liveOnly {
		query.Set("live_only", "true")
	}

	req, err := c.newRequest(ctx, http.MethodGet, "/search/channels", query)
	if err != nil {
		return nil, err
	}

	var resp searchChannelsResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}

	return resp.Data, nil
}

func (c *Client) SearchCategories(ctx context.Context, queryText string, first int) ([]CategorySearchResult, error) {
	queryText = strings.TrimSpace(queryText)
	if queryText == "" {
		return nil, fmt.Errorf("search query is required")
	}

	query := url.Values{}
	query.Set("query", queryText)
	if first > 0 {
		query.Set("first", strconv.Itoa(first))
	}

	req, err := c.newRequest(ctx, http.MethodGet, "/search/categories", query)
	if err != nil {
		return nil, err
	}

	var resp searchCategoriesResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}

	return resp.Data, nil
}
