package api

import (
	"context"
	"net/http"
)

func (c *Client) GetManageableGroups(ctx context.Context) ([]ManageableGroup, error) {
	endpoint, err := buildURL(c.developBaseURL, "/v1/user/groups/canmanagegamesoritems", nil)
	if err != nil {
		return nil, err
	}

	req, err := c.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response ManageableGroupsResponse
	if err := c.do(req, &response); err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) GetUniverses(ctx context.Context) ([]Universe, error) {
	endpoint, err := buildURL(c.developBaseURL, "/v1/user/universes", nil)
	if err != nil {
		return nil, err
	}

	req, err := c.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var response UniversesResponse
	if err := c.do(req, &response); err != nil {
		return nil, err
	}

	return response.Data, nil
}
