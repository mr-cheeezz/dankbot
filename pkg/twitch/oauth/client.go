package oauth

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

const defaultBaseURL = "https://id.twitch.tv/oauth2"

type Client struct {
	httpClient   *http.Client
	baseURL      string
	clientID     string
	clientSecret string
	redirectURI  string
}

func NewClient(httpClient *http.Client, clientID, clientSecret, redirectURI string) *Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 15 * time.Second,
		}
	}

	return &Client{
		httpClient:   httpClient,
		baseURL:      defaultBaseURL,
		clientID:     strings.TrimSpace(clientID),
		clientSecret: strings.TrimSpace(clientSecret),
		redirectURI:  strings.TrimSpace(redirectURI),
	}
}

func (c *Client) RedirectURI() string {
	return c.redirectURI
}

func (c *Client) endpointURL(path string) (string, error) {
	baseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return "", fmt.Errorf("parse twitch oauth base url: %w", err)
	}

	baseURL.Path = strings.TrimRight(baseURL.Path, "/") + path
	return baseURL.String(), nil
}

func (c *Client) newRequest(ctx context.Context, method, path string, query url.Values, body io.Reader) (*http.Request, error) {
	endpoint, err := c.endpointURL(path)
	if err != nil {
		return nil, err
	}

	if query != nil && len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("create oauth request: %w", err)
	}

	return req, nil
}

func (c *Client) do(req *http.Request, out any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute oauth request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return parseAPIError(resp)
	}

	if out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode oauth response: %w", err)
	}

	return nil
}
