package helix

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Subscription struct {
	BroadcasterID    string `json:"broadcaster_id"`
	BroadcasterLogin string `json:"broadcaster_login"`
	BroadcasterName  string `json:"broadcaster_name"`
	GifterID         string `json:"gifter_id"`
	GifterLogin      string `json:"gifter_login"`
	GifterName       string `json:"gifter_name"`
	IsGift           bool   `json:"is_gift"`
	Tier             string `json:"tier"`
	PlanName         string `json:"plan_name"`
	UserID           string `json:"user_id"`
	UserLogin        string `json:"user_login"`
	UserName         string `json:"user_name"`
}

type BroadcasterSubscriptionsPage struct {
	Data       []Subscription `json:"data"`
	Pagination Pagination     `json:"pagination"`
	Total      int            `json:"total"`
	Points     int            `json:"points"`
}

type userSubscriptionResponse struct {
	Data []Subscription `json:"data"`
}

func (c *Client) GetBroadcasterSubscriptions(ctx context.Context, broadcasterID string, userIDs []string, first int, after string) (*BroadcasterSubscriptionsPage, error) {
	broadcasterID = strings.TrimSpace(broadcasterID)
	if broadcasterID == "" {
		return nil, fmt.Errorf("broadcaster id is required")
	}
	if len(userIDs) > 100 {
		return nil, fmt.Errorf("user ids cannot exceed 100")
	}

	query := url.Values{}
	query.Set("broadcaster_id", broadcasterID)

	for _, userID := range userIDs {
		userID = strings.TrimSpace(userID)
		if userID != "" {
			query.Add("user_id", userID)
		}
	}

	if first > 0 {
		query.Set("first", strconv.Itoa(first))
	}
	if after != "" {
		query.Set("after", strings.TrimSpace(after))
	}

	req, err := c.newRequest(ctx, http.MethodGet, "/subscriptions", query)
	if err != nil {
		return nil, err
	}

	var resp BroadcasterSubscriptionsPage
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) CheckUserSubscription(ctx context.Context, broadcasterID, userID string) (*Subscription, error) {
	broadcasterID = strings.TrimSpace(broadcasterID)
	userID = strings.TrimSpace(userID)

	if broadcasterID == "" {
		return nil, fmt.Errorf("broadcaster id is required")
	}
	if userID == "" {
		return nil, fmt.Errorf("user id is required")
	}

	query := url.Values{}
	query.Set("broadcaster_id", broadcasterID)
	query.Set("user_id", userID)

	req, err := c.newRequest(ctx, http.MethodGet, "/subscriptions/user", query)
	if err != nil {
		return nil, err
	}

	var resp userSubscriptionResponse
	if err := c.do(req, &resp); err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
			return nil, nil
		}
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, nil
	}

	return &resp.Data[0], nil
}
