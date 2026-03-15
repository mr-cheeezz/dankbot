package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type EventSubSubscription struct {
	TwitchSubscriptionID string
	SubscriptionType     string
	SubscriptionVersion  string
	Status               string
	Condition            map[string]string
	CallbackURL          string
	TransportMethod      string
	SecretFingerprint    string
	SecretVersion        int
	CreatedAt            time.Time
	UpdatedAt            time.Time
	LastNotificationAt   time.Time
	LastRevokedAt        time.Time
}

type EventSubSubscriptionStore struct {
	client *Client
}

func NewEventSubSubscriptionStore(client *Client) *EventSubSubscriptionStore {
	return &EventSubSubscriptionStore{client: client}
}

func (s *EventSubSubscriptionStore) Save(ctx context.Context, subscription EventSubSubscription) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	conditionJSON, err := json.Marshal(subscription.Condition)
	if err != nil {
		return fmt.Errorf("marshal eventsub condition: %w", err)
	}

	if subscription.TransportMethod == "" {
		subscription.TransportMethod = "webhook"
	}
	if subscription.SecretVersion == 0 {
		subscription.SecretVersion = 1
	}

	query := `
INSERT INTO twitch_eventsub_subscriptions (
	twitch_subscription_id,
	subscription_type,
	subscription_version,
	status,
	condition,
	callback_url,
	transport_method,
	secret_fingerprint,
	secret_version,
	created_at,
	updated_at,
	last_notification_at,
	last_revoked_at
)
VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7, $8, $9, $10, NOW(), $11, $12)
ON CONFLICT (twitch_subscription_id) DO UPDATE SET
	subscription_type = EXCLUDED.subscription_type,
	subscription_version = EXCLUDED.subscription_version,
	status = EXCLUDED.status,
	condition = EXCLUDED.condition,
	callback_url = EXCLUDED.callback_url,
	transport_method = EXCLUDED.transport_method,
	secret_fingerprint = EXCLUDED.secret_fingerprint,
	secret_version = EXCLUDED.secret_version,
	updated_at = NOW(),
	last_notification_at = EXCLUDED.last_notification_at,
	last_revoked_at = EXCLUDED.last_revoked_at
`

	createdAt := subscription.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	_, err = db.ExecContext(
		ctx,
		query,
		subscription.TwitchSubscriptionID,
		subscription.SubscriptionType,
		subscription.SubscriptionVersion,
		subscription.Status,
		string(conditionJSON),
		subscription.CallbackURL,
		subscription.TransportMethod,
		subscription.SecretFingerprint,
		subscription.SecretVersion,
		createdAt,
		nullTime(subscription.LastNotificationAt),
		nullTime(subscription.LastRevokedAt),
	)
	if err != nil {
		return fmt.Errorf("save eventsub subscription %q: %w", subscription.TwitchSubscriptionID, err)
	}

	return nil
}

func (s *EventSubSubscriptionStore) ListByCallback(ctx context.Context, callbackURL string) ([]EventSubSubscription, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	query := `
SELECT
	twitch_subscription_id,
	subscription_type,
	subscription_version,
	status,
	condition,
	callback_url,
	transport_method,
	secret_fingerprint,
	secret_version,
	created_at,
	updated_at,
	last_notification_at,
	last_revoked_at
FROM twitch_eventsub_subscriptions
WHERE callback_url = $1
ORDER BY subscription_type, twitch_subscription_id
`

	rows, err := db.QueryContext(ctx, query, callbackURL)
	if err != nil {
		return nil, fmt.Errorf("list eventsub subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []EventSubSubscription
	for rows.Next() {
		var (
			subscription  EventSubSubscription
			conditionJSON string
			lastNotified  sql.NullTime
			lastRevoked   sql.NullTime
		)

		if err := rows.Scan(
			&subscription.TwitchSubscriptionID,
			&subscription.SubscriptionType,
			&subscription.SubscriptionVersion,
			&subscription.Status,
			&conditionJSON,
			&subscription.CallbackURL,
			&subscription.TransportMethod,
			&subscription.SecretFingerprint,
			&subscription.SecretVersion,
			&subscription.CreatedAt,
			&subscription.UpdatedAt,
			&lastNotified,
			&lastRevoked,
		); err != nil {
			return nil, fmt.Errorf("scan eventsub subscription: %w", err)
		}

		if conditionJSON != "" {
			if err := json.Unmarshal([]byte(conditionJSON), &subscription.Condition); err != nil {
				return nil, fmt.Errorf("unmarshal eventsub condition: %w", err)
			}
		}
		if lastNotified.Valid {
			subscription.LastNotificationAt = lastNotified.Time
		}
		if lastRevoked.Valid {
			subscription.LastRevokedAt = lastRevoked.Time
		}

		subscriptions = append(subscriptions, subscription)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate eventsub subscriptions: %w", err)
	}

	return subscriptions, nil
}

func (s *EventSubSubscriptionStore) ListByTransportMethod(ctx context.Context, transportMethod string) ([]EventSubSubscription, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	query := `
SELECT
	twitch_subscription_id,
	subscription_type,
	subscription_version,
	status,
	condition,
	callback_url,
	transport_method,
	secret_fingerprint,
	secret_version,
	created_at,
	updated_at,
	last_notification_at,
	last_revoked_at
FROM twitch_eventsub_subscriptions
WHERE transport_method = $1
ORDER BY subscription_type, twitch_subscription_id
`

	rows, err := db.QueryContext(ctx, query, transportMethod)
	if err != nil {
		return nil, fmt.Errorf("list eventsub subscriptions by transport method: %w", err)
	}
	defer rows.Close()

	var subscriptions []EventSubSubscription
	for rows.Next() {
		var (
			subscription  EventSubSubscription
			conditionJSON string
			lastNotified  sql.NullTime
			lastRevoked   sql.NullTime
		)

		if err := rows.Scan(
			&subscription.TwitchSubscriptionID,
			&subscription.SubscriptionType,
			&subscription.SubscriptionVersion,
			&subscription.Status,
			&conditionJSON,
			&subscription.CallbackURL,
			&subscription.TransportMethod,
			&subscription.SecretFingerprint,
			&subscription.SecretVersion,
			&subscription.CreatedAt,
			&subscription.UpdatedAt,
			&lastNotified,
			&lastRevoked,
		); err != nil {
			return nil, fmt.Errorf("scan eventsub subscription: %w", err)
		}

		if conditionJSON != "" {
			if err := json.Unmarshal([]byte(conditionJSON), &subscription.Condition); err != nil {
				return nil, fmt.Errorf("unmarshal eventsub condition: %w", err)
			}
		}
		if lastNotified.Valid {
			subscription.LastNotificationAt = lastNotified.Time
		}
		if lastRevoked.Valid {
			subscription.LastRevokedAt = lastRevoked.Time
		}

		subscriptions = append(subscriptions, subscription)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate eventsub subscriptions by transport method: %w", err)
	}

	return subscriptions, nil
}

func (s *EventSubSubscriptionStore) Delete(ctx context.Context, twitchSubscriptionID string) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, `DELETE FROM twitch_eventsub_subscriptions WHERE twitch_subscription_id = $1`, twitchSubscriptionID); err != nil {
		return fmt.Errorf("delete eventsub subscription %q: %w", twitchSubscriptionID, err)
	}

	return nil
}

func (s *EventSubSubscriptionStore) MarkNotification(ctx context.Context, twitchSubscriptionID string, notifiedAt time.Time) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(
		ctx,
		`UPDATE twitch_eventsub_subscriptions SET last_notification_at = $2, updated_at = NOW() WHERE twitch_subscription_id = $1`,
		twitchSubscriptionID,
		notifiedAt,
	); err != nil {
		return fmt.Errorf("mark eventsub notification %q: %w", twitchSubscriptionID, err)
	}

	return nil
}

func (s *EventSubSubscriptionStore) MarkRevoked(ctx context.Context, twitchSubscriptionID, status string, revokedAt time.Time) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(
		ctx,
		`UPDATE twitch_eventsub_subscriptions SET status = $2, last_revoked_at = $3, updated_at = NOW() WHERE twitch_subscription_id = $1`,
		twitchSubscriptionID,
		status,
		revokedAt,
	); err != nil {
		return fmt.Errorf("mark eventsub revoked %q: %w", twitchSubscriptionID, err)
	}

	return nil
}
