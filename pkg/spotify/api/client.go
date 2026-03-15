package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.spotify.com/v1"

type Client struct {
	httpClient  *http.Client
	baseURL     string
	accessToken string
}

func NewClient(httpClient *http.Client, accessToken string) *Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 15 * time.Second,
		}
	}

	return &Client{
		httpClient:  httpClient,
		baseURL:     defaultBaseURL,
		accessToken: strings.TrimSpace(accessToken),
	}
}

func (c *Client) SetAccessToken(accessToken string) {
	c.accessToken = strings.TrimSpace(accessToken)
}

func (c *Client) newRequest(ctx context.Context, method, path string, query url.Values) (*http.Request, error) {
	if strings.TrimSpace(c.accessToken) == "" {
		return nil, fmt.Errorf("spotify access token is required")
	}

	baseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse spotify api base url: %w", err)
	}

	baseURL.Path = strings.TrimRight(baseURL.Path, "/") + path
	if query != nil {
		baseURL.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, baseURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create spotify api request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	return req, nil
}

func (c *Client) do(req *http.Request, out any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute spotify api request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return parseAPIError(resp)
	}

	if out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode spotify api response: %w", err)
	}

	return nil
}
