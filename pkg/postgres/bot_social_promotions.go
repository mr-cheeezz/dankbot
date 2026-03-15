package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type BotSocialPromotion struct {
	ID              int64
	CommandText     string
	IntervalSeconds int
	Enabled         bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
	LastSentAt      time.Time
}

type BotSocialPromotionStore struct {
	client *Client
}

func NewBotSocialPromotionStore(client *Client) *BotSocialPromotionStore {
	return &BotSocialPromotionStore{client: client}
}

func (s *BotSocialPromotionStore) ListEnabled(ctx context.Context) ([]BotSocialPromotion, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(
		ctx,
		`
SELECT
	id,
	command_text,
	interval_seconds,
	enabled,
	created_at,
	updated_at,
	last_sent_at
FROM bot_social_promotions
WHERE enabled = TRUE
ORDER BY id ASC
`,
	)
	if err != nil {
		return nil, fmt.Errorf("list enabled social promotions: %w", err)
	}
	defer rows.Close()

	var items []BotSocialPromotion
	for rows.Next() {
		var item BotSocialPromotion
		var lastSentAt sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.CommandText,
			&item.IntervalSeconds,
			&item.Enabled,
			&item.CreatedAt,
			&item.UpdatedAt,
			&lastSentAt,
		); err != nil {
			return nil, fmt.Errorf("scan social promotion: %w", err)
		}

		if lastSentAt.Valid {
			item.LastSentAt = lastSentAt.Time
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate social promotions: %w", err)
	}

	return items, nil
}

func (s *BotSocialPromotionStore) Save(ctx context.Context, item BotSocialPromotion) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	item.CommandText = strings.TrimSpace(item.CommandText)
	if item.CommandText == "" {
		return fmt.Errorf("social promotion command text is required")
	}
	if item.IntervalSeconds <= 0 {
		return fmt.Errorf("social promotion interval must be greater than 0")
	}

	if item.ID == 0 {
		_, err = db.ExecContext(
			ctx,
			`
INSERT INTO bot_social_promotions (
	command_text,
	interval_seconds,
	enabled,
	created_at,
	updated_at,
	last_sent_at
)
VALUES ($1, $2, $3, NOW(), NOW(), $4)
`,
			item.CommandText,
			item.IntervalSeconds,
			item.Enabled,
			nullTime(item.LastSentAt),
		)
		if err != nil {
			return fmt.Errorf("insert social promotion: %w", err)
		}
		return nil
	}

	_, err = db.ExecContext(
		ctx,
		`
UPDATE bot_social_promotions
SET
	command_text = $2,
	interval_seconds = $3,
	enabled = $4,
	updated_at = NOW()
WHERE id = $1
`,
		item.ID,
		item.CommandText,
		item.IntervalSeconds,
		item.Enabled,
	)
	if err != nil {
		return fmt.Errorf("update social promotion %d: %w", item.ID, err)
	}

	return nil
}

func (s *BotSocialPromotionStore) MarkSent(ctx context.Context, id int64, sentAt time.Time) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(
		ctx,
		`UPDATE bot_social_promotions SET last_sent_at = $2, updated_at = NOW() WHERE id = $1`,
		id,
		sentAt,
	); err != nil {
		return fmt.Errorf("mark social promotion sent %d: %w", id, err)
	}

	return nil
}
