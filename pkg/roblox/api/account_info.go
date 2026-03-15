package api

import (
	"context"
	"net/http"
)

func (c *Client) GetStarCodeAffiliate(ctx context.Context) (*StarCodeAffiliate, error) {
	endpoint, err := buildURL(c.accountInfoBaseURL, "/v1/star-code-affiliates", nil)
	if err != nil {
		return nil, err
	}

	req, err := c.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	var affiliate StarCodeAffiliate
	if err := c.do(req, &affiliate); err != nil {
		return nil, err
	}

	return &affiliate, nil
}
