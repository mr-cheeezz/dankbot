package helix

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Stream struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	UserLogin    string    `json:"user_login"`
	UserName     string    `json:"user_name"`
	GameID       string    `json:"game_id"`
	GameName     string    `json:"game_name"`
	Type         string    `json:"type"`
	Title        string    `json:"title"`
	ViewerCount  int       `json:"viewer_count"`
	StartedAt    time.Time `json:"started_at"`
	Language     string    `json:"language"`
	ThumbnailURL string    `json:"thumbnail_url"`
	IsMature     bool      `json:"is_mature"`
}

type streamsResponse struct {
	Data []Stream `json:"data"`
}

func (c *Client) GetStreamsByUserIDs(ctx context.Context, userIDs []string) ([]Stream, error) {
	query := url.Values{}

	for _, userID := range userIDs {
		userID = strings.TrimSpace(userID)
		if userID != "" {
			query.Add("user_id", userID)
		}
	}

	if len(query["user_id"]) == 0 {
		return nil, fmt.Errorf("at least one user id is required")
	}

	req, err := c.newRequest(ctx, http.MethodGet, "/streams", query)
	if err != nil {
		return nil, err
	}

	var resp streamsResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}

	return resp.Data, nil
}
