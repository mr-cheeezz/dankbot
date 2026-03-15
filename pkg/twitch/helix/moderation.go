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

type BannedUser struct {
	UserID         string     `json:"user_id"`
	UserLogin      string     `json:"user_login"`
	UserName       string     `json:"user_name"`
	ExpiresAt      *time.Time `json:"expires_at"`
	CreatedAt      time.Time  `json:"created_at"`
	Reason         string     `json:"reason"`
	ModeratorID    string     `json:"moderator_id"`
	ModeratorLogin string     `json:"moderator_login"`
	ModeratorName  string     `json:"moderator_name"`
}

type BanUserRequest struct {
	UserID   string `json:"user_id"`
	Duration *int   `json:"duration,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

type Moderator struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
}

type BlockedTerm struct {
	ID        string     `json:"id"`
	Text      string     `json:"text"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type AddBlockedTermRequest struct {
	Text string `json:"text"`
}

type ShieldModeStatus struct {
	IsActive        bool       `json:"is_active"`
	ModeratorID     string     `json:"moderator_id"`
	ModeratorLogin  string     `json:"moderator_login"`
	ModeratorName   string     `json:"moderator_name"`
	LastActivatedAt *time.Time `json:"last_activated_at"`
}

type UpdateShieldModeStatusRequest struct {
	IsActive bool `json:"is_active"`
}

type Warning struct {
	BroadcasterID string `json:"broadcaster_id"`
	UserID        string `json:"user_id"`
	ModeratorID   string `json:"moderator_id"`
	Reason        string `json:"reason"`
}

type WarnChatUserRequest struct {
	UserID string `json:"user_id"`
	Reason string `json:"reason"`
}

type bannedUsersResponse struct {
	Data       []BannedUser `json:"data"`
	Pagination Pagination   `json:"pagination"`
}

type moderatorsResponse struct {
	Data       []Moderator `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

type blockedTermsResponse struct {
	Data       []BlockedTerm `json:"data"`
	Pagination Pagination    `json:"pagination"`
}

type shieldModeResponse struct {
	Data []ShieldModeStatus `json:"data"`
}

type warningsResponse struct {
	Data []Warning `json:"data"`
}

type banUserEnvelope struct {
	Data BanUserRequest `json:"data"`
}

type warnUserEnvelope struct {
	Data WarnChatUserRequest `json:"data"`
}

func (c *Client) GetBannedUsers(ctx context.Context, broadcasterID string, userIDs []string, first int, after string) ([]BannedUser, *Pagination, error) {
	query := url.Values{}
	query.Set("broadcaster_id", strings.TrimSpace(broadcasterID))

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

	req, err := c.newRequest(ctx, http.MethodGet, "/moderation/banned", query)
	if err != nil {
		return nil, nil, err
	}

	var resp bannedUsersResponse
	if err := c.do(req, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Data, &resp.Pagination, nil
}

func (c *Client) BanUser(ctx context.Context, broadcasterID, moderatorID string, body BanUserRequest) (*BannedUser, error) {
	query := url.Values{}
	query.Set("broadcaster_id", strings.TrimSpace(broadcasterID))
	query.Set("moderator_id", strings.TrimSpace(moderatorID))

	req, err := c.newJSONRequest(ctx, http.MethodPost, "/moderation/bans", query, banUserEnvelope{Data: body})
	if err != nil {
		return nil, err
	}

	var resp bannedUsersResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("twitch returned no banned user")
	}

	return &resp.Data[0], nil
}

func (c *Client) UnbanUser(ctx context.Context, broadcasterID, moderatorID, userID string) error {
	query := url.Values{}
	query.Set("broadcaster_id", strings.TrimSpace(broadcasterID))
	query.Set("moderator_id", strings.TrimSpace(moderatorID))
	query.Set("user_id", strings.TrimSpace(userID))

	req, err := c.newRequest(ctx, http.MethodDelete, "/moderation/bans", query)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) DeleteChatMessages(ctx context.Context, broadcasterID, moderatorID, messageID string) error {
	query := url.Values{}
	query.Set("broadcaster_id", strings.TrimSpace(broadcasterID))
	query.Set("moderator_id", strings.TrimSpace(moderatorID))
	if messageID != "" {
		query.Set("message_id", strings.TrimSpace(messageID))
	}

	req, err := c.newRequest(ctx, http.MethodDelete, "/moderation/chat", query)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) GetModerators(ctx context.Context, broadcasterID string, userIDs []string, first int, after string) ([]Moderator, *Pagination, error) {
	query := url.Values{}
	query.Set("broadcaster_id", strings.TrimSpace(broadcasterID))

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

	req, err := c.newRequest(ctx, http.MethodGet, "/moderation/moderators", query)
	if err != nil {
		return nil, nil, err
	}

	var resp moderatorsResponse
	if err := c.do(req, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Data, &resp.Pagination, nil
}

func (c *Client) AddChannelModerator(ctx context.Context, broadcasterID, userID string) error {
	query := url.Values{}
	query.Set("broadcaster_id", strings.TrimSpace(broadcasterID))
	query.Set("user_id", strings.TrimSpace(userID))

	req, err := c.newRequest(ctx, http.MethodPost, "/moderation/moderators", query)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) RemoveChannelModerator(ctx context.Context, broadcasterID, userID string) error {
	query := url.Values{}
	query.Set("broadcaster_id", strings.TrimSpace(broadcasterID))
	query.Set("user_id", strings.TrimSpace(userID))

	req, err := c.newRequest(ctx, http.MethodDelete, "/moderation/moderators", query)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) GetBlockedTerms(ctx context.Context, broadcasterID, moderatorID string, first int, after string) ([]BlockedTerm, *Pagination, error) {
	query := url.Values{}
	query.Set("broadcaster_id", strings.TrimSpace(broadcasterID))
	query.Set("moderator_id", strings.TrimSpace(moderatorID))

	if first > 0 {
		query.Set("first", strconv.Itoa(first))
	}
	if after != "" {
		query.Set("after", strings.TrimSpace(after))
	}

	req, err := c.newRequest(ctx, http.MethodGet, "/moderation/blocked_terms", query)
	if err != nil {
		return nil, nil, err
	}

	var resp blockedTermsResponse
	if err := c.do(req, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Data, &resp.Pagination, nil
}

func (c *Client) AddBlockedTerm(ctx context.Context, broadcasterID, moderatorID string, body AddBlockedTermRequest) (*BlockedTerm, error) {
	query := url.Values{}
	query.Set("broadcaster_id", strings.TrimSpace(broadcasterID))
	query.Set("moderator_id", strings.TrimSpace(moderatorID))

	req, err := c.newJSONRequest(ctx, http.MethodPost, "/moderation/blocked_terms", query, body)
	if err != nil {
		return nil, err
	}

	var resp blockedTermsResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("twitch returned no blocked term")
	}

	return &resp.Data[0], nil
}

func (c *Client) RemoveBlockedTerm(ctx context.Context, broadcasterID, moderatorID, termID string) error {
	query := url.Values{}
	query.Set("broadcaster_id", strings.TrimSpace(broadcasterID))
	query.Set("moderator_id", strings.TrimSpace(moderatorID))
	query.Set("id", strings.TrimSpace(termID))

	req, err := c.newRequest(ctx, http.MethodDelete, "/moderation/blocked_terms", query)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) GetShieldModeStatus(ctx context.Context, broadcasterID, moderatorID string) (*ShieldModeStatus, error) {
	query := url.Values{}
	query.Set("broadcaster_id", strings.TrimSpace(broadcasterID))
	query.Set("moderator_id", strings.TrimSpace(moderatorID))

	req, err := c.newRequest(ctx, http.MethodGet, "/moderation/shield_mode", query)
	if err != nil {
		return nil, err
	}

	var resp shieldModeResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("twitch returned no shield mode status")
	}

	return &resp.Data[0], nil
}

func (c *Client) UpdateShieldModeStatus(ctx context.Context, broadcasterID, moderatorID string, isActive bool) (*ShieldModeStatus, error) {
	query := url.Values{}
	query.Set("broadcaster_id", strings.TrimSpace(broadcasterID))
	query.Set("moderator_id", strings.TrimSpace(moderatorID))

	req, err := c.newJSONRequest(ctx, http.MethodPut, "/moderation/shield_mode", query, UpdateShieldModeStatusRequest{
		IsActive: isActive,
	})
	if err != nil {
		return nil, err
	}

	var resp shieldModeResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("twitch returned no updated shield mode status")
	}

	return &resp.Data[0], nil
}

func (c *Client) WarnChatUser(ctx context.Context, broadcasterID, moderatorID string, body WarnChatUserRequest) (*Warning, error) {
	query := url.Values{}
	query.Set("broadcaster_id", strings.TrimSpace(broadcasterID))
	query.Set("moderator_id", strings.TrimSpace(moderatorID))

	req, err := c.newJSONRequest(ctx, http.MethodPost, "/moderation/warnings", query, warnUserEnvelope{Data: body})
	if err != nil {
		return nil, err
	}

	var resp warningsResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("twitch returned no warning")
	}

	return &resp.Data[0], nil
}
