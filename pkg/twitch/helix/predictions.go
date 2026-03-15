package helix

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type PredictionOutcome struct {
	ID            string         `json:"id"`
	Title         string         `json:"title"`
	Users         int            `json:"users"`
	ChannelPoints int            `json:"channel_points"`
	TopPredictors []TopPredictor `json:"top_predictors"`
	Color         string         `json:"color"`
}

type TopPredictor struct {
	UserID            string `json:"user_id"`
	UserLogin         string `json:"user_login"`
	UserName          string `json:"user_name"`
	ChannelPointsUsed int    `json:"channel_points_used"`
	ChannelPointsWon  int    `json:"channel_points_won"`
}

type Prediction struct {
	ID               string              `json:"id"`
	BroadcasterID    string              `json:"broadcaster_id"`
	BroadcasterLogin string              `json:"broadcaster_login"`
	BroadcasterName  string              `json:"broadcaster_name"`
	Title            string              `json:"title"`
	WinningOutcomeID string              `json:"winning_outcome_id"`
	Outcomes         []PredictionOutcome `json:"outcomes"`
	PredictionWindow int                 `json:"prediction_window"`
	Status           string              `json:"status"`
	CreatedAt        time.Time           `json:"created_at"`
	EndedAt          *time.Time          `json:"ended_at"`
	LockedAt         *time.Time          `json:"locked_at"`
}

type PredictionOutcomeInput struct {
	Title string `json:"title"`
}

type CreatePredictionRequest struct {
	BroadcasterID    string                   `json:"broadcaster_id"`
	Title            string                   `json:"title"`
	Outcomes         []PredictionOutcomeInput `json:"outcomes"`
	PredictionWindow int                      `json:"prediction_window"`
}

type EndPredictionRequest struct {
	BroadcasterID    string `json:"broadcaster_id"`
	ID               string `json:"id"`
	Status           string `json:"status"`
	WinningOutcomeID string `json:"winning_outcome_id,omitempty"`
}

type predictionsResponse struct {
	Data       []Prediction `json:"data"`
	Pagination Pagination   `json:"pagination"`
}

func (c *Client) GetPredictions(ctx context.Context, broadcasterID string, ids []string, first int, after string) ([]Prediction, *Pagination, error) {
	query := url.Values{}
	query.Set("broadcaster_id", strings.TrimSpace(broadcasterID))

	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id != "" {
			query.Add("id", id)
		}
	}

	if first > 0 {
		query.Set("first", strconv.Itoa(first))
	}
	if after != "" {
		query.Set("after", strings.TrimSpace(after))
	}

	req, err := c.newRequest(ctx, http.MethodGet, "/predictions", query)
	if err != nil {
		return nil, nil, err
	}

	var resp predictionsResponse
	if err := c.do(req, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Data, &resp.Pagination, nil
}

func (c *Client) CreatePrediction(ctx context.Context, body CreatePredictionRequest) (*Prediction, error) {
	req, err := c.newJSONRequest(ctx, http.MethodPost, "/predictions", nil, body)
	if err != nil {
		return nil, err
	}

	var resp predictionsResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("twitch returned no prediction")
	}

	return &resp.Data[0], nil
}

func (c *Client) EndPrediction(ctx context.Context, body EndPredictionRequest) (*Prediction, error) {
	req, err := c.newJSONRequest(ctx, http.MethodPatch, "/predictions", nil, body)
	if err != nil {
		return nil, err
	}

	var resp predictionsResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("twitch returned no ended prediction")
	}

	return &resp.Data[0], nil
}
