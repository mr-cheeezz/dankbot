package eventsub

import "time"

type AdBreakBeginEvent struct {
	BroadcasterUserID    string    `json:"broadcaster_user_id"`
	BroadcasterUserLogin string    `json:"broadcaster_user_login"`
	BroadcasterUserName  string    `json:"broadcaster_user_name"`
	RequesterUserID      string    `json:"requester_user_id"`
	RequesterUserLogin   string    `json:"requester_user_login"`
	RequesterUserName    string    `json:"requester_user_name"`
	DurationSeconds      int       `json:"duration_seconds"`
	StartedAt            time.Time `json:"started_at"`
	IsAutomatic          bool      `json:"is_automatic"`
}

type PollChoice struct {
	ID                 string `json:"id"`
	Title              string `json:"title"`
	BitsVotes          int    `json:"bits_votes"`
	ChannelPointsVotes int    `json:"channel_points_votes"`
	Votes              int    `json:"votes"`
}

type ChannelPointsVoting struct {
	IsEnabled     bool `json:"is_enabled"`
	AmountPerVote int  `json:"amount_per_vote"`
}

type PollEvent struct {
	ID                   string              `json:"id"`
	BroadcasterUserID    string              `json:"broadcaster_user_id"`
	BroadcasterUserLogin string              `json:"broadcaster_user_login"`
	BroadcasterUserName  string              `json:"broadcaster_user_name"`
	Title                string              `json:"title"`
	Choices              []PollChoice        `json:"choices"`
	ChannelPointsVoting  ChannelPointsVoting `json:"channel_points_voting"`
	StartedAt            time.Time           `json:"started_at"`
	EndsAt               *time.Time          `json:"ends_at"`
	EndedAt              time.Time           `json:"ended_at"`
	Status               string              `json:"status"`
}

type RedemptionReward struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Cost   int    `json:"cost"`
	Prompt string `json:"prompt"`
}

type ChannelPointRedemptionEvent struct {
	ID                   string           `json:"id"`
	BroadcasterUserID    string           `json:"broadcaster_user_id"`
	BroadcasterUserLogin string           `json:"broadcaster_user_login"`
	BroadcasterUserName  string           `json:"broadcaster_user_name"`
	UserID               string           `json:"user_id"`
	UserLogin            string           `json:"user_login"`
	UserName             string           `json:"user_name"`
	UserInput            string           `json:"user_input"`
	Status               string           `json:"status"`
	Reward               RedemptionReward `json:"reward"`
	RedeemedAt           time.Time        `json:"redeemed_at"`
}

type StreamStatusEvent struct {
	ID                   string    `json:"id"`
	BroadcasterUserID    string    `json:"broadcaster_user_id"`
	BroadcasterUserLogin string    `json:"broadcaster_user_login"`
	BroadcasterUserName  string    `json:"broadcaster_user_name"`
	Type                 string    `json:"type"`
	StartedAt            time.Time `json:"started_at"`
}

type SubscriptionEvent struct {
	UserID               string `json:"user_id"`
	UserLogin            string `json:"user_login"`
	UserName             string `json:"user_name"`
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
	Tier                 string `json:"tier"`
	IsGift               bool   `json:"is_gift"`
}

type SubscriptionGiftEvent struct {
	UserID               string `json:"user_id"`
	UserLogin            string `json:"user_login"`
	UserName             string `json:"user_name"`
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
	Total                int    `json:"total"`
	Tier                 string `json:"tier"`
	CumulativeTotal      *int   `json:"cumulative_total"`
	IsAnonymous          bool   `json:"is_anonymous"`
}

type CheerEvent struct {
	IsAnonymous          bool   `json:"is_anonymous"`
	UserID               string `json:"user_id"`
	UserLogin            string `json:"user_login"`
	UserName             string `json:"user_name"`
	BroadcasterUserID    string `json:"broadcaster_user_id"`
	BroadcasterUserLogin string `json:"broadcaster_user_login"`
	BroadcasterUserName  string `json:"broadcaster_user_name"`
	Message              string `json:"message"`
	Bits                 int    `json:"bits"`
}

type PredictionOutcome struct {
	ID            string                `json:"id"`
	Title         string                `json:"title"`
	Color         string                `json:"color"`
	Users         int                   `json:"users"`
	ChannelPoints int                   `json:"channel_points"`
	TopPredictors []PredictionPredictor `json:"top_predictors"`
}

type PredictionPredictor struct {
	UserID            string `json:"user_id"`
	UserLogin         string `json:"user_login"`
	UserName          string `json:"user_name"`
	ChannelPointsUsed int    `json:"channel_points_used"`
	ChannelPointsWon  int    `json:"channel_points_won"`
}

type PredictionEvent struct {
	ID                   string              `json:"id"`
	BroadcasterUserID    string              `json:"broadcaster_user_id"`
	BroadcasterUserLogin string              `json:"broadcaster_user_login"`
	BroadcasterUserName  string              `json:"broadcaster_user_name"`
	Title                string              `json:"title"`
	Outcomes             []PredictionOutcome `json:"outcomes"`
	StartedAt            time.Time           `json:"started_at"`
	LocksAt              *time.Time          `json:"locks_at"`
	EndedAt              *time.Time          `json:"ended_at"`
	Status               string              `json:"status"`
	WinningOutcomeID     string              `json:"winning_outcome_id"`
}

type HypeTrainContribution struct {
	UserID    string `json:"user_id"`
	UserLogin string `json:"user_login"`
	UserName  string `json:"user_name"`
	Type      string `json:"type"`
	Total     int    `json:"total"`
}

type HypeTrainEvent struct {
	ID                   string                  `json:"id"`
	BroadcasterUserID    string                  `json:"broadcaster_user_id"`
	BroadcasterUserLogin string                  `json:"broadcaster_user_login"`
	BroadcasterUserName  string                  `json:"broadcaster_user_name"`
	Total                int                     `json:"total"`
	Progress             int                     `json:"progress"`
	Goal                 int                     `json:"goal"`
	TopContributions     []HypeTrainContribution `json:"top_contributions"`
	LastContribution     *HypeTrainContribution  `json:"last_contribution"`
	Level                int                     `json:"level"`
	StartedAt            *time.Time              `json:"started_at"`
	ExpiresAt            *time.Time              `json:"expires_at"`
	CooldownEndsAt       *time.Time              `json:"cooldown_ends_at"`
}
