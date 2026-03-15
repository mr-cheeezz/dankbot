package api

import (
	"context"
	"net/http"
)

func (c *Client) TriggerAlert(ctx context.Context, payload AlertRequest) error {
	return c.doJSON(ctx, http.MethodPost, "/alerts", nil, payload, nil)
}

func (c *Client) SendTestAlert(ctx context.Context, payload AlertRequest) error {
	return c.doJSON(ctx, http.MethodPost, "/alerts/send_test_alert", nil, payload, nil)
}
