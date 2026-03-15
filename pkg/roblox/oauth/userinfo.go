package oauth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

func (c *Client) UserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return nil, fmt.Errorf("access token is required")
	}

	req, err := c.newRequest(ctx, http.MethodGet, "/v1/userinfo", nil, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	var userInfo UserInfo
	if err := c.do(req, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}
