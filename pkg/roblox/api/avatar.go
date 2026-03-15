package api

import (
	"context"
	"fmt"
	"net/http"
)

func (c *Client) GetAuthenticatedAvatar(ctx context.Context) (*Avatar, error) {
	endpoint, err := buildURL(c.avatarBaseURL, "/v2/avatar/avatar", nil)
	if err != nil {
		return nil, err
	}

	req, err := c.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var avatar Avatar
	if err := c.do(req, &avatar); err != nil {
		return nil, err
	}

	return &avatar, nil
}

func (c *Client) GetCurrentlyWearing(ctx context.Context, userID int64) ([]int64, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("user id must be greater than 0")
	}

	endpoint, err := buildURL(c.avatarBaseURL, fmt.Sprintf("/v1/users/%d/currently-wearing", userID), nil)
	if err != nil {
		return nil, err
	}

	req, err := c.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response CurrentlyWearingResponse
	if err := c.do(req, &response); err != nil {
		return nil, err
	}

	return response.AssetIDs, nil
}
