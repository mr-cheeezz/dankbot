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

const defaultBaseURL = "https://api.streamelements.com/kappa/v2"

type Client struct {
	httpClient  *http.Client
	baseURL     string
	accessToken string
}

func NewClient(httpClient *http.Client, accessToken string) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}

	return &Client{
		httpClient:  httpClient,
		baseURL:     defaultBaseURL,
		accessToken: strings.TrimSpace(accessToken),
	}
}

func (c *Client) newRequest(ctx context.Context, method, path string) (*http.Request, error) {
	if strings.TrimSpace(c.accessToken) == "" {
		return nil, fmt.Errorf("streamelements access token is required")
	}

	baseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse streamelements api base url: %w", err)
	}

	baseURL.Path = strings.TrimRight(baseURL.Path, "/") + path

	req, err := http.NewRequestWithContext(ctx, method, baseURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create streamelements api request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	return req, nil
}

func (c *Client) do(req *http.Request, out any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute streamelements api request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return fmt.Errorf("streamelements api request failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	if out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode streamelements api response: %w", err)
	}

	return nil
}
