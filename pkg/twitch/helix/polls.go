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

type PollChoice struct {
	ID                 string `json:"id"`
	Title              string `json:"title"`
	Votes              int    `json:"votes"`
	ChannelPointsVotes int    `json:"channel_points_votes"`
	BitsVotes          int    `json:"bits_votes"`
}

type Poll struct {
	ID                         string       `json:"id"`
	BroadcasterID              string       `json:"broadcaster_id"`
	BroadcasterName            string       `json:"broadcaster_name"`
	BroadcasterLogin           string       `json:"broadcaster_login"`
	Title                      string       `json:"title"`
	Choices                    []PollChoice `json:"choices"`
	BitsVotingEnabled          bool         `json:"bits_voting_enabled"`
	BitsPerVote                int          `json:"bits_per_vote"`
	ChannelPointsVotingEnabled bool         `json:"channel_points_voting_enabled"`
	ChannelPointsPerVote       int          `json:"channel_points_per_vote"`
	Status                     string       `json:"status"`
	Duration                   int          `json:"duration"`
	StartedAt                  time.Time    `json:"started_at"`
	EndedAt                    *time.Time   `json:"ended_at"`
}

type PollChoiceInput struct {
	Title string `json:"title"`
}

type CreatePollRequest struct {
	BroadcasterID              string            `json:"broadcaster_id"`
	Title                      string            `json:"title"`
	Choices                    []PollChoiceInput `json:"choices"`
	Duration                   int               `json:"duration"`
	BitsVotingEnabled          *bool             `json:"bits_voting_enabled,omitempty"`
	BitsPerVote                *int              `json:"bits_per_vote,omitempty"`
	ChannelPointsVotingEnabled *bool             `json:"channel_points_voting_enabled,omitempty"`
	ChannelPointsPerVote       *int              `json:"channel_points_per_vote,omitempty"`
}

type EndPollRequest struct {
	BroadcasterID string `json:"broadcaster_id"`
	ID            string `json:"id"`
	Status        string `json:"status"`
}

type pollsResponse struct {
	Data       []Poll     `json:"data"`
	Pagination Pagination `json:"pagination"`
}

func (c *Client) GetPolls(ctx context.Context, broadcasterID string, ids []string, first int, after string) ([]Poll, *Pagination, error) {
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

	req, err := c.newRequest(ctx, http.MethodGet, "/polls", query)
	if err != nil {
		return nil, nil, err
	}

	var resp pollsResponse
	if err := c.do(req, &resp); err != nil {
		return nil, nil, err
	}

	return resp.Data, &resp.Pagination, nil
}

func (c *Client) CreatePoll(ctx context.Context, body CreatePollRequest) (*Poll, error) {
	req, err := c.newJSONRequest(ctx, http.MethodPost, "/polls", nil, body)
	if err != nil {
		return nil, err
	}

	var resp pollsResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("twitch returned no poll")
	}

	return &resp.Data[0], nil
}

func (c *Client) EndPoll(ctx context.Context, body EndPollRequest) (*Poll, error) {
	req, err := c.newJSONRequest(ctx, http.MethodPatch, "/polls", nil, body)
	if err != nil {
		return nil, err
	}

	var resp pollsResponse
	if err := c.do(req, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("twitch returned no ended poll")
	}

	return &resp.Data[0], nil
}
