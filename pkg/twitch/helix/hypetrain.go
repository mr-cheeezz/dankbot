package helix

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HypeTrainContribution struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
	Type      string `json:"type"`
	Total     int    `json:"total"`
}

type SharedTrainParticipant struct {
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
}

type CurrentHypeTrain struct {
	ID                      string                   `json:"id"`
	BroadcasterUserID       string                   `json:"broadcaster_user_id"`
	BroadcasterUserLogin    string                   `json:"broadcaster_user_login"`
	BroadcasterUserName     string                   `json:"broadcaster_user_name"`
	Level                   int                      `json:"level"`
	Total                   int                      `json:"total"`
	Progress                int                      `json:"progress"`
	Goal                    int                      `json:"goal"`
	TopContributions        []HypeTrainContribution  `json:"top_contributions"`
	SharedTrainParticipants []SharedTrainParticipant `json:"shared_train_participants"`
	StartedAt               *time.Time               `json:"started_at"`
	ExpiresAt               *time.Time               `json:"expires_at"`
	Type                    string                   `json:"type"`
}

type HypeTrainRecord struct {
	Level      int        `json:"level"`
	Total      int        `json:"total"`
	AchievedAt *time.Time `json:"achieved_at"`
}

type HypeTrainStatus struct {
	Current           *CurrentHypeTrain `json:"current"`
	AllTimeHigh       *HypeTrainRecord  `json:"all_time_high"`
	SharedAllTimeHigh *HypeTrainRecord  `json:"shared_all_time_high"`
}

type hypeTrainStatusResponse struct {
	Data []HypeTrainStatus `json:"data"`
}

func (c *Client) GetHypeTrainStatus(ctx context.Context, broadcasterID string) (*HypeTrainStatus, error) {
	query := url.Values{}
	query.Set("broadcaster_id", strings.TrimSpace(broadcasterID))

	req, err := c.newRequest(ctx, http.MethodGet, "/hypetrain/status", query)
	if err != nil {
		return nil, err
	}

	var resp hypeTrainStatusResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("twitch returned no hype train status")
	}

	return &resp.Data[0], nil
}
