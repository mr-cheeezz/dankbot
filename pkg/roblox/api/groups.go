package api

import (
	"context"
	"fmt"
	"net/http"
)

func (c *Client) GetGroupRoles(ctx context.Context, userID int64) ([]GroupRole, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("user id must be greater than 0")
	}

	endpoint, err := buildURL(c.groupsBaseURL, fmt.Sprintf("/v2/users/%d/groups/roles", userID), nil)
	if err != nil {
		return nil, err
	}

	req, err := c.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response GroupRolesResponse
	if err := c.do(req, &response); err != nil {
		return nil, err
	}

	return response.Data, nil
}

func (c *Client) GetPrimaryGroupRole(ctx context.Context, userID int64) (*PrimaryGroupRole, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("user id must be greater than 0")
	}

	endpoint, err := buildURL(c.groupsBaseURL, fmt.Sprintf("/v1/users/%d/groups/primary/role", userID), nil)
	if err != nil {
		return nil, err
	}

	req, err := c.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response PrimaryGroupRole
	if err := c.do(req, &response); err != nil {
		return nil, err
	}

	return &response, nil
}
