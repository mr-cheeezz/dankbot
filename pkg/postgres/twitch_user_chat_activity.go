package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type TwitchUserChatActivity struct {
	UserID       string
	UserLogin    string
	DisplayName  string
	MessageCount int64
	LastSeenAt   time.Time
	LastChatAt   time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type TwitchUserChatActivityStore struct {
	client *Client
}

func NewTwitchUserChatActivityStore(client *Client) *TwitchUserChatActivityStore {
	return &TwitchUserChatActivityStore{client: client}
}

func (s *TwitchUserChatActivityStore) Touch(ctx context.Context, userID, userLogin, displayName string, at time.Time) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	userID = strings.TrimSpace(userID)
	userLogin = strings.TrimSpace(strings.ToLower(strings.TrimPrefix(userLogin, "@")))
	displayName = strings.TrimSpace(displayName)
	if userID == "" || userLogin == "" {
		return nil
	}

	if at.IsZero() {
		at = time.Now().UTC()
	} else {
		at = at.UTC()
	}

	_, err = db.ExecContext(
		ctx,
		`
INSERT INTO twitch_user_chat_activity (
	user_id,
	user_login,
	display_name,
	message_count,
	last_seen_at,
	last_chat_at,
	created_at,
	updated_at
)
VALUES ($1, $2, $3, 1, $4, $4, NOW(), NOW())
ON CONFLICT (user_id) DO UPDATE SET
	user_login = EXCLUDED.user_login,
	display_name = CASE
		WHEN EXCLUDED.display_name = '' THEN twitch_user_chat_activity.display_name
		ELSE EXCLUDED.display_name
	END,
	message_count = twitch_user_chat_activity.message_count + 1,
	last_seen_at = EXCLUDED.last_seen_at,
	last_chat_at = EXCLUDED.last_chat_at,
	updated_at = NOW()
`,
		userID,
		userLogin,
		displayName,
		at,
	)
	if err != nil {
		return fmt.Errorf("touch twitch user chat activity %q: %w", userID, err)
	}
	return nil
}

func (s *TwitchUserChatActivityStore) GetByUserID(ctx context.Context, userID string) (*TwitchUserChatActivity, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, nil
	}

	var row TwitchUserChatActivity
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	user_id,
	user_login,
	display_name,
	message_count,
	last_seen_at,
	last_chat_at,
	created_at,
	updated_at
FROM twitch_user_chat_activity
WHERE user_id = $1
`,
		userID,
	).Scan(
		&row.UserID,
		&row.UserLogin,
		&row.DisplayName,
		&row.MessageCount,
		&row.LastSeenAt,
		&row.LastChatAt,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get twitch user chat activity %q: %w", userID, err)
	}
	return &row, nil
}
