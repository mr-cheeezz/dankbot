package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

func (c *Client) GetTrack(ctx context.Context, trackID string) (*Track, error) {
	trackID = strings.TrimSpace(trackID)
	if trackID == "" {
		return nil, fmt.Errorf("spotify track id is required")
	}

	req, err := c.newRequest(ctx, http.MethodGet, "/tracks/"+trackID, nil)
	if err != nil {
		return nil, err
	}

	var track Track
	if err := c.do(req, &track); err != nil {
		return nil, err
	}

	return &track, nil
}
