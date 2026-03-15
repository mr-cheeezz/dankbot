package eventsub

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.twitch.tv/helix"

type Client struct {
	httpClient  *http.Client
	baseURL     string
	clientID    string
	accessToken string
}

func NewClient(httpClient *http.Client, clientID, accessToken string) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}

	return &Client{
		httpClient:  httpClient,
		baseURL:     defaultBaseURL,
		clientID:    strings.TrimSpace(clientID),
		accessToken: strings.TrimSpace(accessToken),
	}
}

func (c *Client) ListSubscriptions(ctx context.Context) ([]Subscription, error) {
	var subscriptions []Subscription
	var cursor string

	for {
		query := url.Values{}
		if cursor != "" {
			query.Set("after", cursor)
		}

		req, err := c.newRequest(ctx, http.MethodGet, "/eventsub/subscriptions", query, nil)
		if err != nil {
			return nil, err
		}

		var resp subscriptionsResponse
		if err := c.do(req, &resp); err != nil {
			return nil, err
		}

		subscriptions = append(subscriptions, resp.Data...)
		if resp.Pagination.Cursor == "" {
			return subscriptions, nil
		}

		cursor = resp.Pagination.Cursor
	}
}

func (c *Client) CreateSubscription(ctx context.Context, subscription DesiredSubscription, transport Transport) (*Subscription, error) {
	reqBody := createSubscriptionRequest{
		Type:      subscription.Type,
		Version:   subscription.Version,
		Condition: subscription.Condition,
		Transport: transport,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(reqBody); err != nil {
		return nil, fmt.Errorf("encode eventsub subscription request: %w", err)
	}

	req, err := c.newRequest(ctx, http.MethodPost, "/eventsub/subscriptions", nil, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	var resp subscriptionsResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("twitch returned no eventsub subscription")
	}

	return &resp.Data[0], nil
}

func (c *Client) DeleteSubscription(ctx context.Context, subscriptionID string) error {
	query := url.Values{}
	query.Set("id", strings.TrimSpace(subscriptionID))

	req, err := c.newRequest(ctx, http.MethodDelete, "/eventsub/subscriptions", query, nil)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) newRequest(ctx context.Context, method, path string, query url.Values, body io.Reader) (*http.Request, error) {
	if c.clientID == "" {
		return nil, fmt.Errorf("eventsub client id is required")
	}
	if c.accessToken == "" {
		return nil, fmt.Errorf("eventsub access token is required")
	}

	baseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse eventsub base url: %w", err)
	}

	baseURL.Path = strings.TrimRight(baseURL.Path, "/") + path
	if query != nil {
		baseURL.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, baseURL.String(), body)
	if err != nil {
		return nil, fmt.Errorf("create eventsub request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Client-Id", c.clientID)
	return req, nil
}

func (c *Client) do(req *http.Request, out any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute eventsub request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		if readErr != nil {
			return fmt.Errorf("read eventsub error response: %w", readErr)
		}
		return fmt.Errorf("eventsub request failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	if out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode eventsub response: %w", err)
	}

	return nil
}
