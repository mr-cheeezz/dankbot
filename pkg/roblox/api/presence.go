package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) GetPresences(ctx context.Context, userIDs []int64) ([]UserPresence, error) {
	if len(userIDs) == 0 {
		return nil, fmt.Errorf("at least one user id is required")
	}

	body, err := json.Marshal(PresenceRequest{UserIDs: userIDs})
	if err != nil {
		return nil, fmt.Errorf("marshal roblox presence request: %w", err)
	}

	endpoint, err := buildURL(c.presenceURL, "/v1/presence/users", nil)
	if err != nil {
		return nil, err
	}

	req, err := c.newRequest(ctx, http.MethodPost, endpoint, body)
	if err != nil {
		return nil, err
	}

	var response PresenceResponse
	if err := c.do(req, &response); err != nil {
		return nil, err
	}

	return response.UserPresences, nil
}
