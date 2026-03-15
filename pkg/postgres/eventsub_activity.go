package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type PollChoiceSnapshot struct {
	ChoiceID           string
	Title              string
	Votes              int
	ChannelPointsVotes int
	BitsVotes          int
}

type PollEventSnapshot struct {
	TwitchSubscriptionID string
	EventType            string
	PollID               string
	BroadcasterUserID    string
	BroadcasterUserLogin string
	BroadcasterUserName  string
	Title                string
	Status               string
	StartedAt            time.Time
	EndedAt              time.Time
	RawEvent             json.RawMessage
	Choices              []PollChoiceSnapshot
}

type ChannelPointRedemption struct {
	RedemptionID         string
	TwitchSubscriptionID string
	BroadcasterUserID    string
	BroadcasterUserLogin string
	BroadcasterUserName  string
	UserID               string
	UserLogin            string
	UserName             string
	UserInput            string
	Status               string
	RedeemedAt           time.Time
	RewardID             string
	RewardTitle          string
	RewardCost           int
	RewardPrompt         string
	RawEvent             json.RawMessage
}

type EventSubActivityStore struct {
	client *Client
}

type UserRedemptionRewardSummary struct {
	RewardTitle      string
	RedemptionCount  int
	TotalPointsSpent int
}

type UserRedemptionActivity struct {
	RewardTitle string
	RewardCost  int
	Status      string
	UserInput   string
	RedeemedAt  time.Time
}

type UserRedemptionStats struct {
	RedemptionCount  int
	TotalPointsSpent int
	LastRedeemedAt   time.Time
	TopRewards       []UserRedemptionRewardSummary
	RecentActivity   []UserRedemptionActivity
}

func NewEventSubActivityStore(client *Client) *EventSubActivityStore {
	return &EventSubActivityStore{client: client}
}

func (s *EventSubActivityStore) SavePollEvent(ctx context.Context, event PollEventSnapshot) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin poll event transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	query := `
INSERT INTO twitch_poll_events (
	twitch_subscription_id,
	event_type,
	poll_id,
	broadcaster_user_id,
	broadcaster_user_login,
	broadcaster_user_name,
	title,
	status,
	started_at,
	ended_at,
	raw_event
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::jsonb)
RETURNING id
`

	var pollEventID int64
	if scanErr := tx.QueryRowContext(
		ctx,
		query,
		event.TwitchSubscriptionID,
		event.EventType,
		event.PollID,
		event.BroadcasterUserID,
		event.BroadcasterUserLogin,
		event.BroadcasterUserName,
		event.Title,
		event.Status,
		nullTime(event.StartedAt),
		nullTime(event.EndedAt),
		string(event.RawEvent),
	).Scan(&pollEventID); scanErr != nil {
		err = fmt.Errorf("insert poll event snapshot: %w", scanErr)
		return err
	}

	for _, choice := range event.Choices {
		if _, execErr := tx.ExecContext(
			ctx,
			`INSERT INTO twitch_poll_event_choices (poll_event_id, choice_id, title, votes, channel_points_votes, bits_votes) VALUES ($1, $2, $3, $4, $5, $6)`,
			pollEventID,
			choice.ChoiceID,
			choice.Title,
			choice.Votes,
			choice.ChannelPointsVotes,
			choice.BitsVotes,
		); execErr != nil {
			err = fmt.Errorf("insert poll choice snapshot: %w", execErr)
			return err
		}
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("commit poll event transaction: %w", commitErr)
	}

	return nil
}

func (s *EventSubActivityStore) SaveChannelPointRedemption(ctx context.Context, redemption ChannelPointRedemption) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	query := `
INSERT INTO twitch_channel_point_redemptions (
	redemption_id,
	twitch_subscription_id,
	broadcaster_user_id,
	broadcaster_user_login,
	broadcaster_user_name,
	user_id,
	user_login,
	user_name,
	user_input,
	status,
	redeemed_at,
	reward_id,
	reward_title,
	reward_cost,
	reward_prompt,
	raw_event,
	created_at,
	updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16::jsonb, NOW(), NOW())
ON CONFLICT (redemption_id) DO UPDATE SET
	twitch_subscription_id = EXCLUDED.twitch_subscription_id,
	broadcaster_user_id = EXCLUDED.broadcaster_user_id,
	broadcaster_user_login = EXCLUDED.broadcaster_user_login,
	broadcaster_user_name = EXCLUDED.broadcaster_user_name,
	user_id = EXCLUDED.user_id,
	user_login = EXCLUDED.user_login,
	user_name = EXCLUDED.user_name,
	user_input = EXCLUDED.user_input,
	status = EXCLUDED.status,
	redeemed_at = EXCLUDED.redeemed_at,
	reward_id = EXCLUDED.reward_id,
	reward_title = EXCLUDED.reward_title,
	reward_cost = EXCLUDED.reward_cost,
	reward_prompt = EXCLUDED.reward_prompt,
	raw_event = EXCLUDED.raw_event,
	updated_at = NOW()
`

	if _, err := db.ExecContext(
		ctx,
		query,
		redemption.RedemptionID,
		redemption.TwitchSubscriptionID,
		redemption.BroadcasterUserID,
		redemption.BroadcasterUserLogin,
		redemption.BroadcasterUserName,
		redemption.UserID,
		redemption.UserLogin,
		redemption.UserName,
		redemption.UserInput,
		redemption.Status,
		nullTime(redemption.RedeemedAt),
		redemption.RewardID,
		redemption.RewardTitle,
		redemption.RewardCost,
		redemption.RewardPrompt,
		string(redemption.RawEvent),
	); err != nil {
		return fmt.Errorf("save channel point redemption %q: %w", redemption.RedemptionID, err)
	}

	return nil
}

func (s *EventSubActivityStore) GetUserRedemptionStats(ctx context.Context, userID string, topLimit, recentLimit int) (UserRedemptionStats, error) {
	var stats UserRedemptionStats

	db, err := s.client.DB(ctx)
	if err != nil {
		return stats, err
	}

	if topLimit <= 0 {
		topLimit = 3
	}
	if recentLimit <= 0 {
		recentLimit = 6
	}

	var lastRedeemedAt sql.NullTime
	if err := db.QueryRowContext(
		ctx,
		`SELECT COUNT(*), COALESCE(SUM(reward_cost), 0), MAX(COALESCE(redeemed_at, created_at))
		 FROM twitch_channel_point_redemptions
		 WHERE user_id = $1`,
		userID,
	).Scan(&stats.RedemptionCount, &stats.TotalPointsSpent, &lastRedeemedAt); err != nil {
		return stats, fmt.Errorf("get user redemption summary %q: %w", userID, err)
	}
	if lastRedeemedAt.Valid {
		stats.LastRedeemedAt = lastRedeemedAt.Time
	}

	topRows, err := db.QueryContext(
		ctx,
		`SELECT reward_title, COUNT(*) AS redemption_count, COALESCE(SUM(reward_cost), 0) AS total_points_spent
		 FROM twitch_channel_point_redemptions
		 WHERE user_id = $1
		 GROUP BY reward_title
		 ORDER BY redemption_count DESC, total_points_spent DESC, reward_title ASC
		 LIMIT $2`,
		userID,
		topLimit,
	)
	if err != nil {
		return stats, fmt.Errorf("list top redemption rewards %q: %w", userID, err)
	}
	defer topRows.Close()

	for topRows.Next() {
		var item UserRedemptionRewardSummary
		if err := topRows.Scan(&item.RewardTitle, &item.RedemptionCount, &item.TotalPointsSpent); err != nil {
			return stats, fmt.Errorf("scan top redemption reward %q: %w", userID, err)
		}
		stats.TopRewards = append(stats.TopRewards, item)
	}
	if err := topRows.Err(); err != nil {
		return stats, fmt.Errorf("iterate top redemption rewards %q: %w", userID, err)
	}

	recentRows, err := db.QueryContext(
		ctx,
		`SELECT reward_title, reward_cost, status, user_input, COALESCE(redeemed_at, created_at) AS activity_at
		 FROM twitch_channel_point_redemptions
		 WHERE user_id = $1
		 ORDER BY activity_at DESC
		 LIMIT $2`,
		userID,
		recentLimit,
	)
	if err != nil {
		return stats, fmt.Errorf("list recent redemption activity %q: %w", userID, err)
	}
	defer recentRows.Close()

	for recentRows.Next() {
		var item UserRedemptionActivity
		if err := recentRows.Scan(&item.RewardTitle, &item.RewardCost, &item.Status, &item.UserInput, &item.RedeemedAt); err != nil {
			return stats, fmt.Errorf("scan recent redemption activity %q: %w", userID, err)
		}
		stats.RecentActivity = append(stats.RecentActivity, item)
	}
	if err := recentRows.Err(); err != nil {
		return stats, fmt.Errorf("iterate recent redemption activity %q: %w", userID, err)
	}

	return stats, nil
}
