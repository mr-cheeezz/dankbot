package oauth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

func (c *Client) ValidateToken(ctx context.Context, accessToken string) (*ValidateResponse, error) {
	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return nil, fmt.Errorf("access token is required")
	}

	req, err := c.newRequest(ctx, http.MethodGet, "/validate", nil, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "OAuth "+accessToken)

	var validation ValidateResponse
	if err := c.do(req, &validation); err != nil {
		return nil, err
	}

	return &validation, nil
}
