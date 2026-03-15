package helix

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Channel struct {
	BroadcasterID    string `json:"broadcaster_id"`
	BroadcasterLogin string `json:"broadcaster_login"`
	BroadcasterName  string `json:"broadcaster_name"`
	GameID           string `json:"game_id"`
	GameName         string `json:"game_name"`
	Title            string `json:"title"`
	Delay            int    `json:"delay"`
	Language         string `json:"broadcaster_language"`
	IsBrandedContent bool   `json:"is_branded_content"`
}

type channelsResponse struct {
	Data []Channel `json:"data"`
}

type UpdateChannelInformationRequest struct {
	Title *string `json:"title,omitempty"`
}

func (c *Client) GetChannelsByBroadcasterIDs(ctx context.Context, broadcasterIDs []string) ([]Channel, error) {
	query := url.Values{}

	for _, broadcasterID := range broadcasterIDs {
		broadcasterID = strings.TrimSpace(broadcasterID)
		if broadcasterID != "" {
			query.Add("broadcaster_id", broadcasterID)
		}
	}

	if len(query["broadcaster_id"]) == 0 {
		return nil, fmt.Errorf("at least one broadcaster id is required")
	}

	req, err := c.newRequest(ctx, http.MethodGet, "/channels", query)
	if err != nil {
		return nil, err
	}

	var resp channelsResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}

	return resp.Data, nil
}

func (c *Client) UpdateChannelInformation(ctx context.Context, broadcasterID string, request UpdateChannelInformationRequest) error {
	broadcasterID = strings.TrimSpace(broadcasterID)
	if broadcasterID == "" {
		return fmt.Errorf("broadcaster id is required")
	}

	query := url.Values{}
	query.Set("broadcaster_id", broadcasterID)

	req, err := c.newJSONRequest(ctx, http.MethodPatch, "/channels", query, request)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}
