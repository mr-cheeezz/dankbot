package helix

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Chatter struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
}

type ChatSettings struct {
	BroadcasterID                 string `json:"broadcaster_id"`
	SlowMode                      bool   `json:"slow_mode"`
	SlowModeWaitTime              int    `json:"slow_mode_wait_time"`
	FollowerMode                  bool   `json:"follower_mode"`
	FollowerModeDuration          int    `json:"follower_mode_duration"`
	SubscriberMode                bool   `json:"subscriber_mode"`
	EmoteMode                     bool   `json:"emote_mode"`
	UniqueChatMode                bool   `json:"unique_chat_mode"`
	NonModeratorChatDelay         bool   `json:"non_moderator_chat_delay"`
	NonModeratorChatDelayDuration int    `json:"non_moderator_chat_delay_duration"`
}

type UpdateChatSettingsRequest struct {
	EmoteMode                     *bool `json:"emote_mode,omitempty"`
	FollowerMode                  *bool `json:"follower_mode,omitempty"`
	FollowerModeDuration          *int  `json:"follower_mode_duration,omitempty"`
	NonModeratorChatDelay         *bool `json:"non_moderator_chat_delay,omitempty"`
	NonModeratorChatDelayDuration *int  `json:"non_moderator_chat_delay_duration,omitempty"`
	SlowMode                      *bool `json:"slow_mode,omitempty"`
	SlowModeWaitTime              *int  `json:"slow_mode_wait_time,omitempty"`
	SubscriberMode                *bool `json:"subscriber_mode,omitempty"`
	UniqueChatMode                *bool `json:"unique_chat_mode,omitempty"`
}

type SendChatAnnouncementRequest struct {
	Message string `json:"message"`
	Color   string `json:"color,omitempty"`
}

type SendChatMessageRequest struct {
	BroadcasterID        string `json:"broadcaster_id"`
	SenderID             string `json:"sender_id"`
	Message              string `json:"message"`
	ReplyParentMessageID string `json:"reply_parent_message_id,omitempty"`
	ForSourceOnly        *bool  `json:"for_source_only,omitempty"`
}

type SentChatMessage struct {
	MessageID  string      `json:"message_id"`
	IsSent     bool        `json:"is_sent"`
	DropReason *DropReason `json:"drop_reason"`
}

type DropReason struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type chattersResponse struct {
	Data       []Chatter  `json:"data"`
	Pagination Pagination `json:"pagination"`
	Total      int        `json:"total"`
}

type chatSettingsResponse struct {
	Data []ChatSettings `json:"data"`
}

type sendChatMessageResponse struct {
	Data []SentChatMessage `json:"data"`
}

func (c *Client) GetChatters(ctx context.Context, broadcasterID, moderatorID string, first int, after string) ([]Chatter, *Pagination, int, error) {
	query := url.Values{}
	query.Set("broadcaster_id", strings.TrimSpace(broadcasterID))
	query.Set("moderator_id", strings.TrimSpace(moderatorID))

	if first > 0 {
		query.Set("first", strconv.Itoa(first))
	}
	if after != "" {
		query.Set("after", strings.TrimSpace(after))
	}

	req, err := c.newRequest(ctx, http.MethodGet, "/chat/chatters", query)
	if err != nil {
		return nil, nil, 0, err
	}

	var resp chattersResponse
	if err := c.do(req, &resp); err != nil {
		return nil, nil, 0, err
	}

	return resp.Data, &resp.Pagination, resp.Total, nil
}

func (c *Client) GetChatSettings(ctx context.Context, broadcasterID, moderatorID string) (*ChatSettings, error) {
	query := url.Values{}
	query.Set("broadcaster_id", strings.TrimSpace(broadcasterID))
	if moderatorID != "" {
		query.Set("moderator_id", strings.TrimSpace(moderatorID))
	}

	req, err := c.newRequest(ctx, http.MethodGet, "/chat/settings", query)
	if err != nil {
		return nil, err
	}

	var resp chatSettingsResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("twitch returned no chat settings")
	}

	return &resp.Data[0], nil
}

func (c *Client) UpdateChatSettings(ctx context.Context, broadcasterID, moderatorID string, body UpdateChatSettingsRequest) (*ChatSettings, error) {
	query := url.Values{}
	query.Set("broadcaster_id", strings.TrimSpace(broadcasterID))
	query.Set("moderator_id", strings.TrimSpace(moderatorID))

	req, err := c.newJSONRequest(ctx, http.MethodPatch, "/chat/settings", query, body)
	if err != nil {
		return nil, err
	}

	var resp chatSettingsResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("twitch returned no updated chat settings")
	}

	return &resp.Data[0], nil
}

func (c *Client) SendChatAnnouncement(ctx context.Context, broadcasterID, moderatorID string, body SendChatAnnouncementRequest) error {
	query := url.Values{}
	query.Set("broadcaster_id", strings.TrimSpace(broadcasterID))
	query.Set("moderator_id", strings.TrimSpace(moderatorID))

	req, err := c.newJSONRequest(ctx, http.MethodPost, "/chat/announcements", query, body)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) SendShoutout(ctx context.Context, broadcasterID, moderatorID, toBroadcasterID string) error {
	query := url.Values{}
	query.Set("broadcaster_id", strings.TrimSpace(broadcasterID))
	query.Set("moderator_id", strings.TrimSpace(moderatorID))
	query.Set("to_broadcaster_id", strings.TrimSpace(toBroadcasterID))

	req, err := c.newRequest(ctx, http.MethodPost, "/chat/shoutouts", query)
	if err != nil {
		return err
	}

	return c.do(req, nil)
}

func (c *Client) SendChatMessage(ctx context.Context, body SendChatMessageRequest) (*SentChatMessage, error) {
	req, err := c.newJSONRequest(ctx, http.MethodPost, "/chat/messages", nil, body)
	if err != nil {
		return nil, err
	}

	var resp sendChatMessageResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("twitch returned no chat message response")
	}

	return &resp.Data[0], nil
}
