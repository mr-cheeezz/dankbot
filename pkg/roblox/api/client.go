package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	defaultUsersBaseURL       = "https://users.roblox.com"
	defaultFriendsBaseURL     = "https://friends.roblox.com"
	defaultPresenceBaseURL    = "https://presence.roblox.com"
	defaultGroupsBaseURL      = "https://groups.roblox.com"
	defaultDevelopBaseURL     = "https://develop.roblox.com"
	defaultAvatarBaseURL      = "https://avatar.roblox.com"
	defaultAccountInfoBaseURL = "https://accountinformation.roblox.com"
)

type Client struct {
	httpClient         *http.Client
	cookie             string
	mu                 sync.Mutex
	csrfToken          string
	usersBaseURL       string
	friendsURL         string
	presenceURL        string
	groupsBaseURL      string
	developBaseURL     string
	avatarBaseURL      string
	accountInfoBaseURL string
}

func NewClient(httpClient *http.Client, cookie string) *Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 15 * time.Second,
		}
	}

	return &Client{
		httpClient:         httpClient,
		cookie:             strings.TrimSpace(cookie),
		usersBaseURL:       defaultUsersBaseURL,
		friendsURL:         defaultFriendsBaseURL,
		presenceURL:        defaultPresenceBaseURL,
		groupsBaseURL:      defaultGroupsBaseURL,
		developBaseURL:     defaultDevelopBaseURL,
		avatarBaseURL:      defaultAvatarBaseURL,
		accountInfoBaseURL: defaultAccountInfoBaseURL,
	}
}

func (c *Client) newRequest(ctx context.Context, method, rawURL string, body []byte) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, rawURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create roblox api request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Cookie", ".ROBLOSECURITY="+c.cookie)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if csrfToken := c.getCSRFToken(); csrfToken != "" {
		req.Header.Set("x-csrf-token", csrfToken)
	}

	return req, nil
}

func (c *Client) getCSRFToken() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.csrfToken
}

func (c *Client) setCSRFToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.csrfToken = strings.TrimSpace(token)
}

func (c *Client) do(req *http.Request, out any) error {
	if strings.TrimSpace(c.cookie) == "" {
		return fmt.Errorf("roblox cookie is required")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute roblox api request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		if token := strings.TrimSpace(resp.Header.Get("x-csrf-token")); token != "" && req.Method != http.MethodGet && req.Method != http.MethodHead {
			c.setCSRFToken(token)

			retryBody, err := cloneRequestBody(req)
			if err != nil {
				return err
			}

			retryReq, err := c.newRequest(req.Context(), req.Method, req.URL.String(), retryBody)
			if err != nil {
				return err
			}

			resp.Body.Close()
			resp, err = c.httpClient.Do(retryReq)
			if err != nil {
				return fmt.Errorf("retry roblox api request: %w", err)
			}
			defer resp.Body.Close()
		}
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return parseAPIError(resp)
	}

	if out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode roblox api response: %w", err)
	}

	return nil
}

func cloneRequestBody(req *http.Request) ([]byte, error) {
	if req.Body == nil {
		return nil, nil
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("read roblox request body: %w", err)
	}

	req.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}

func buildURL(baseURL, path string, query url.Values) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return "", fmt.Errorf("parse roblox api base url: %w", err)
	}

	parsed.Path = strings.TrimRight(parsed.Path, "/") + path
	if query != nil {
		parsed.RawQuery = query.Encode()
	}

	return parsed.String(), nil
}
