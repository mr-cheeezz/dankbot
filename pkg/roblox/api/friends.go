package api

import (
	"context"
	"fmt"
	"net/http"
)

func (c *Client) GetFriends(ctx context.Context, userID int64) ([]Friend, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("user id must be greater than 0")
	}

	endpoint, err := buildURL(c.friendsURL, fmt.Sprintf("/v1/users/%d/friends", userID), nil)
	if err != nil {
		return nil, err
	}

	req, err := c.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response FriendsResponse
	if err := c.do(req, &response); err != nil {
		return nil, err
	}

	return response.Data, nil
}
