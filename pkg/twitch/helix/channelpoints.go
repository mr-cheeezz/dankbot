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

type RewardImage struct {
	URL1x string `json:"url_1x"`
	URL2x string `json:"url_2x"`
	URL4x string `json:"url_4x"`
}

type MaxPerStreamSetting struct {
	IsEnabled    bool `json:"is_enabled"`
	MaxPerStream int  `json:"max_per_stream"`
}

type MaxPerUserPerStreamSetting struct {
	IsEnabled           bool `json:"is_enabled"`
	MaxPerUserPerStream int  `json:"max_per_user_per_stream"`
}

type GlobalCooldownSetting struct {
	IsEnabled             bool `json:"is_enabled"`
	GlobalCooldownSeconds int  `json:"global_cooldown_seconds"`
}

type CustomReward struct {
	BroadcasterID                     string                     `json:"broadcaster_id"`
	BroadcasterLogin                  string                     `json:"broadcaster_login"`
	BroadcasterName                   string                     `json:"broadcaster_name"`
	ID                                string                     `json:"id"`
	Title                             string                     `json:"title"`
	Prompt                            string                     `json:"prompt"`
	Cost                              int                        `json:"cost"`
	BackgroundColor                   string                     `json:"background_color"`
	IsEnabled                         bool                       `json:"is_enabled"`
	IsUserInputRequired               bool                       `json:"is_user_input_required"`
	IsPaused                          bool                       `json:"is_paused"`
	IsInStock                         bool                       `json:"is_in_stock"`
	ShouldRedemptionsSkipRequestQueue bool                       `json:"should_redemptions_skip_request_queue"`
	Image                             *RewardImage               `json:"image"`
	DefaultImage                      RewardImage                `json:"default_image"`
	MaxPerStreamSetting               MaxPerStreamSetting        `json:"max_per_stream_setting"`
	MaxPerUserPerStreamSetting        MaxPerUserPerStreamSetting `json:"max_per_user_per_stream_setting"`
	GlobalCooldownSetting             GlobalCooldownSetting      `json:"global_cooldown_setting"`
	RedemptionsRedeemedCurrentStream  *int                       `json:"redemptions_redeemed_current_stream"`
	CooldownExpiresAt                 *time.Time                 `json:"cooldown_expires_at"`
}

type CreateCustomRewardRequest struct {
	Title                             string `json:"title"`
	Cost                              int    `json:"cost"`
	Prompt                            string `json:"prompt,omitempty"`
	BackgroundColor                   string `json:"background_color,omitempty"`
	IsEnabled                         *bool  `json:"is_enabled,omitempty"`
	IsUserInputRequired               *bool  `json:"is_user_input_required,omitempty"`
	IsMaxPerStreamEnabled             *bool  `json:"is_max_per_stream_enabled,omitempty"`
	MaxPerStream                      *int   `json:"max_per_stream,omitempty"`
	IsMaxPerUserPerStreamEnabled      *bool  `json:"is_max_per_user_per_stream_enabled,omitempty"`
	MaxPerUserPerStream               *int   `json:"max_per_user_per_stream,omitempty"`
	IsGlobalCooldownEnabled           *bool  `json:"is_global_cooldown_enabled,omitempty"`
	GlobalCooldownSeconds             *int   `json:"global_cooldown_seconds,omitempty"`
	ShouldRedemptionsSkipRequestQueue *bool  `json:"should_redemptions_skip_request_queue,omitempty"`
}

type UpdateCustomRewardRequest struct {
	Title                             *string `json:"title,omitempty"`
	Cost                              *int    `json:"cost,omitempty"`
	Prompt                            *string `json:"prompt,omitempty"`
	BackgroundColor                   *string `json:"background_color,omitempty"`
	IsEnabled                         *bool   `json:"is_enabled,omitempty"`
	IsUserInputRequired               *bool   `json:"is_user_input_required,omitempty"`
	IsPaused                          *bool   `json:"is_paused,omitempty"`
	IsInStock                         *bool   `json:"is_in_stock,omitempty"`
	ShouldRedemptionsSkipRequestQueue *bool   `json:"should_redemptions_skip_request_queue,omitempty"`
	IsMaxPerStreamEnabled             *bool   `json:"is_max_per_stream_enabled,omitempty"`
	MaxPerStream                      *int    `json:"max_per_stream,omitempty"`
	IsMaxPerUserPerStreamEnabled      *bool   `json:"is_max_per_user_per_stream_enabled,omitempty"`
	MaxPerUserPerStream               *int    `json:"max_per_user_per_stream,omitempty"`
	IsGlobalCooldownEnabled           *bool   `json:"is_global_cooldown_enabled,omitempty"`
	GlobalCooldownSeconds             *int    `json:"global_cooldown_seconds,omitempty"`
}

type RedemptionReward struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Prompt string `json:"prompt"`
	Cost   int    `json:"cost"`
}

type CustomRewardRedemption struct {
	ID               string           `json:"id"`
	BroadcasterID    string           `json:"broadcaster_id"`
	BroadcasterLogin string           `json:"broadcaster_login"`
	BroadcasterName  string           `json:"broadcaster_name"`
	UserID           string           `json:"user_id"`
	UserLogin        string           `json:"user_login"`
	UserName         string           `json:"user_name"`
	UserInput        string           `json:"user_input"`
	Status           string           `json:"status"`
	RedeemedAt       time.Time        `json:"redeemed_at"`
	Reward           RedemptionReward `json:"reward"`
}

type UpdateRedemptionStatusRequest struct {
	Status string `json:"status"`
}

type customRewardsResponse struct {
	Data []CustomReward `json:"data"`
}

type customRewardRedemptionsResponse struct {
	Data       []CustomRewardRedemption `json:"data"`
	Pagination Pagination               `json:"pagination"`
}

func (c *Client) GetCustomRewards(ctx context.Context, broadcasterID string, rewardIDs []string, onlyManageable bool) ([]CustomReward, error) {
	broadcasterID = strings.TrimSpace(broadcasterID)
	if broadcasterID == "" {
		return nil, fmt.Errorf("broadcaster id is required")
	}

	query := url.Values{}
	query.Set("broadcaster_id", broadcasterID)

	for _, rewardID := range rewardIDs {
		rewardID = strings.TrimSpace(rewardID)
		if rewardID != "" {
			query.Add("id", rewardID)
		}
	}

	if onlyManageable {
		query.Set("only_manageable_rewards", strconv.FormatBool(true))
	}

	req, err := c.newRequest(ctx, http.MethodGet, "/channel_points/custom_rewards", query)
	if err != nil {
		return nil, err
	}

	var resp customRewardsResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}

	return resp.Data, nil
}

func (c *Client) CreateCustomReward(ctx context.Context, broadcasterID string, body CreateCustomRewardRequest) (*CustomReward, error) {
	query := url.Values{}
	query.Set("broadcaster_id", strings.TrimSpace(broadcasterID))

	req, err := c.newJSONRequest(ctx, http.MethodPost, "/channel_points/custom_rewards", query, body)
	if err != nil {
		return nil, err
	}

	var resp customRewardsResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("twitch returned no custom reward")
	}

	return &resp.Data[0], nil
}

func (c *Client) UpdateCustomReward(ctx context.Context, broadcasterID, rewardID string, body UpdateCustomRewardRequest) (*CustomReward, error) {
	query := url.Values{}
	query.Set("broadcaster_id", strings.TrimSpace(broadcasterID))
	query.Set("id", strings.TrimSpace(rewardID))

	req, err := c.newJSONRequest(ctx, http.MethodPatch, "/channel_points/custom_rewards", query, body)
	if err != nil {
		return nil, err
	}

	var resp customRewardsResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("twitch returned no updated custom reward")
	}

	return &resp.Data[0], nil
}

func (c *Client) DeleteCustomReward(ctx context.Context, broadcasterID, rewardID string) error {
	query := url.Values{}
	query.Set("broadcaster_id", strings.TrimSpace(broadcasterID))
	query.Set("id", strings.TrimSpace(rewardID))

	req, err := c.newRequest(ctx, http.MethodDelete, "/channel_points/custom_rewards", query)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) GetCustomRewardRedemptions(ctx context.Context, broadcasterID, rewardID string, rewardStatus string, redemptionIDs []string, sortOrder string, first int, after string) ([]CustomRewardRedemption, *Pagination, error) {
	query := url.Values{}
	query.Set("broadcaster_id", strings.TrimSpace(broadcasterID))
	query.Set("reward_id", strings.TrimSpace(rewardID))

	if rewardStatus != "" {
		query.Set("status", strings.TrimSpace(rewardStatus))
	}
	if sortOrder != "" {
		query.Set("sort", strings.TrimSpace(sortOrder))
	}
	if first > 0 {
		query.Set("first", strconv.Itoa(first))
	}
	if after != "" {
		query.Set("after", strings.TrimSpace(after))
	}

	for _, redemptionID := range redemptionIDs {
		redemptionID = strings.TrimSpace(redemptionID)
		if redemptionID != "" {
			query.Add("id", redemptionID)
		}
	}

	req, err := c.newRequest(ctx, http.MethodGet, "/channel_points/custom_rewards/redemptions", query)
	if err != nil {
		return nil, nil, err
	}

	var resp customRewardRedemptionsResponse
	if err := c.do(req, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Data, &resp.Pagination, nil
}

func (c *Client) UpdateRedemptionStatus(ctx context.Context, broadcasterID, rewardID string, redemptionIDs []string, status string) ([]CustomRewardRedemption, error) {
	query := url.Values{}
	query.Set("broadcaster_id", strings.TrimSpace(broadcasterID))
	query.Set("reward_id", strings.TrimSpace(rewardID))

	for _, redemptionID := range redemptionIDs {
		redemptionID = strings.TrimSpace(redemptionID)
		if redemptionID != "" {
			query.Add("id", redemptionID)
		}
	}

	req, err := c.newJSONRequest(ctx, http.MethodPatch, "/channel_points/custom_rewards/redemptions", query, UpdateRedemptionStatusRequest{
		Status: strings.TrimSpace(status),
	})
	if err != nil {
		return nil, err
	}

	var resp customRewardRedemptionsResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}

	return resp.Data, nil
}
