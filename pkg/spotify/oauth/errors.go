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
	if e.ErrorCode != "" {
		return fmt.Sprintf("spotify oauth request failed with status %d (%s): %s", e.StatusCode, e.ErrorCode, e.Message)
	}

	return fmt.Sprintf("spotify oauth request failed with status %d: %s", e.StatusCode, e.Message)
}

type errorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func parseAPIError(resp *http.Response) error {
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read spotify oauth error response: %w", err)
	}

	var oauthErr errorResponse
	if err := json.Unmarshal(body, &oauthErr); err == nil {
		if oauthErr.Error != "" || oauthErr.ErrorDescription != "" {
			return &APIError{
				StatusCode: resp.StatusCode,
				ErrorCode:  oauthErr.Error,
				Message:    oauthErr.ErrorDescription,
			}
		}
	}

	return &APIError{
		StatusCode: resp.StatusCode,
		Message:    strings.TrimSpace(string(body)),
	}
}
