package oauth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type APIError struct {
	StatusCode int
	ErrorCode  string
	Message    string
}

func (e *APIError) Error() string {
	if e.ErrorCode == "" {
		return fmt.Sprintf("twitch oauth request failed with status %d: %s", e.StatusCode, e.Message)
	}

	return fmt.Sprintf("twitch oauth request failed with status %d (%s): %s", e.StatusCode, e.ErrorCode, e.Message)
}

type errorResponse struct {
	StatusCode int    `json:"status"`
	ErrorCode  string `json:"error"`
	Message    string `json:"message"`
}

func parseAPIError(resp *http.Response) error {
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read oauth error response: %w", err)
	}

	var oauthErr errorResponse
	if err := json.Unmarshal(body, &oauthErr); err == nil && oauthErr.Message != "" {
		return &APIError{
			StatusCode: resp.StatusCode,
			ErrorCode:  oauthErr.ErrorCode,
			Message:    oauthErr.Message,
		}
	}

	return &APIError{
		StatusCode: resp.StatusCode,
		Message:    strings.TrimSpace(string(body)),
	}
}
