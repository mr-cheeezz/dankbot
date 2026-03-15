package helix

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type APIError struct {
	StatusCode int
	ErrorType  string
	Message    string
}

func (e *APIError) Error() string {
	if e.ErrorType == "" {
		return fmt.Sprintf("twitch helix request failed with status %d: %s", e.StatusCode, e.Message)
	}

	return fmt.Sprintf("twitch helix request failed with status %d (%s): %s", e.StatusCode, e.ErrorType, e.Message)
}

type errorResponse struct {
	ErrorType string `json:"error"`
	Status    int    `json:"status"`
	Message   string `json:"message"`
}

func (c *Client) newRequest(ctx context.Context, method, endpoint string, query url.Values) (*http.Request, error) {
	if c.clientID == "" {
		return nil, fmt.Errorf("helix client id is required")
	}
	if c.accessToken == "" {
		return nil, fmt.Errorf("helix access token is required")
	}

	baseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse helix base url: %w", err)
	}

	baseURL.Path = strings.TrimRight(baseURL.Path, "/") + endpoint
	baseURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, method, baseURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create helix request: %w", err)
	}

	req.Header.Set("Client-Id", c.clientID)
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Accept", "application/json")

	return req, nil
}

func (c *Client) do(req *http.Request, out any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute helix request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		if readErr != nil {
			return fmt.Errorf("read helix error response: %w", readErr)
		}

		var apiErr errorResponse
		if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Message != "" {
			return &APIError{
				StatusCode: resp.StatusCode,
				ErrorType:  apiErr.ErrorType,
				Message:    apiErr.Message,
			}
		}

		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    strings.TrimSpace(string(body)),
		}
	}

	if out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode helix response: %w", err)
	}

	return nil
}

func (c *Client) newJSONRequest(ctx context.Context, method, endpoint string, query url.Values, body any) (*http.Request, error) {
	if c.clientID == "" {
		return nil, fmt.Errorf("helix client id is required")
	}
	if c.accessToken == "" {
		return nil, fmt.Errorf("helix access token is required")
	}

	var bodyReader io.Reader
	if body != nil {
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, fmt.Errorf("encode helix request body: %w", err)
		}
		bodyReader = &buf
	}

	baseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse helix base url: %w", err)
	}

	baseURL.Path = strings.TrimRight(baseURL.Path, "/") + endpoint
	if query != nil {
		baseURL.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, baseURL.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create helix request: %w", err)
	}

	req.Header.Set("Client-Id", c.clientID)
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}
