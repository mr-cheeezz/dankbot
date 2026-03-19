package helix

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type VIP struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
}

type vipsResponse struct {
	Data       []VIP      `json:"data"`
	Pagination Pagination `json:"pagination"`
}

func (c *Client) GetVIPs(ctx context.Context, broadcasterID string, userIDs []string, first int, after string) ([]VIP, *Pagination, error) {
	query := url.Values{}
	query.Set("broadcaster_id", strings.TrimSpace(broadcasterID))

	for _, userID := range userIDs {
		userID = strings.TrimSpace(userID)
		if userID != "" {
			query.Add("user_id", userID)
		}
	}

	if first > 0 {
		query.Set("first", strconv.Itoa(first))
	}
	if strings.TrimSpace(after) != "" {
		query.Set("after", strings.TrimSpace(after))
	}

	req, err := c.newRequest(ctx, http.MethodGet, "/channels/vips", query)
	if err != nil {
		return nil, nil, err
	}

	var resp vipsResponse
	if err := c.do(req, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Data, &resp.Pagination, nil
}
