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

type ChannelFollower struct {
	UserID     string    `json:"user_id"`
	UserLogin  string    `json:"user_login"`
	UserName   string    `json:"user_name"`
	FollowedAt time.Time `json:"followed_at"`
}

type channelFollowersResponse struct {
	Total      int               `json:"total"`
	Data       []ChannelFollower `json:"data"`
	Pagination Pagination        `json:"pagination"`
}

func (c *Client) GetChannelFollowers(ctx context.Context, broadcasterID, moderatorID string, first int, after string) ([]ChannelFollower, int, *Pagination, error) {
	query := url.Values{}
	query.Set("broadcaster_id", strings.TrimSpace(broadcasterID))
	query.Set("moderator_id", strings.TrimSpace(moderatorID))

	if first > 0 {
		query.Set("first", strconv.Itoa(first))
	}
	if after != "" {
		query.Set("after", strings.TrimSpace(after))
	}

	req, err := c.newRequest(ctx, http.MethodGet, "/channels/followers", query)
	if err != nil {
		return nil, 0, nil, err
	}

	var resp channelFollowersResponse
	if err := c.do(req, &resp); err != nil {
		return nil, 0, nil, err
	}

	return resp.Data, resp.Total, &resp.Pagination, nil
}

func (c *Client) GetRecentChannelFollowers(ctx context.Context, broadcasterID, moderatorID string, limit int) ([]ChannelFollower, int, error) {
	if limit <= 0 {
		return nil, 0, fmt.Errorf("limit must be greater than zero")
	}

	const maxPerPage = 100

	items := make([]ChannelFollower, 0, limit)
	total := 0
	after := ""

	for len(items) < limit {
		first := limit - len(items)
		if first > maxPerPage {
			first = maxPerPage
		}

		page, pageTotal, pagination, err := c.GetChannelFollowers(ctx, broadcasterID, moderatorID, first, after)
		if err != nil {
			return nil, 0, err
		}
		total = pageTotal
		items = append(items, page...)

		if pagination == nil || strings.TrimSpace(pagination.Cursor) == "" || len(page) == 0 {
			break
		}

		after = pagination.Cursor
	}

	if len(items) > limit {
		items = items[:limit]
	}

	return items, total, nil
}
