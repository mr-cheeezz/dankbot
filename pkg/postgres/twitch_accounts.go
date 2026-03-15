package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type TwitchAccountKind string

const (
	TwitchAccountKindStreamer TwitchAccountKind = "streamer"
	TwitchAccountKindBot      TwitchAccountKind = "bot"
)

type TwitchAccount struct {
	Kind            TwitchAccountKind
	TwitchUserID    string
	Login           string
	DisplayName     string
	AccessToken     string
	RefreshToken    string
	Scopes          []string
	TokenType       string
	ExpiresAt       time.Time
	LastValidatedAt time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type TwitchAccountStore struct {
	client *Client
}

func NewTwitchAccountStore(client *Client) *TwitchAccountStore {
	return &TwitchAccountStore{client: client}
}

func (s *TwitchAccountStore) Save(ctx context.Context, account TwitchAccount) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	scopesJSON, err := json.Marshal(account.Scopes)
	if err != nil {
		return fmt.Errorf("marshal twitch scopes: %w", err)
	}

	if account.DisplayName == "" {
		account.DisplayName = account.Login
	}

	query := `
INSERT INTO twitch_accounts (
	kind,
	twitch_user_id,
	login,
	display_name,
	access_token,
	refresh_token,
	scopes,
	token_type,
	expires_at,
	last_validated_at,
	created_at,
	updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, $9, $10, NOW(), NOW())
ON CONFLICT (kind) DO UPDATE SET
	twitch_user_id = EXCLUDED.twitch_user_id,
	login = EXCLUDED.login,
	display_name = EXCLUDED.display_name,
	access_token = EXCLUDED.access_token,
	refresh_token = EXCLUDED.refresh_token,
	scopes = EXCLUDED.scopes,
	token_type = EXCLUDED.token_type,
	expires_at = EXCLUDED.expires_at,
	last_validated_at = EXCLUDED.last_validated_at,
	updated_at = NOW()
`

	_, err = db.ExecContext(
		ctx,
		query,
		account.Kind,
		account.TwitchUserID,
		account.Login,
		account.DisplayName,
		account.AccessToken,
		account.RefreshToken,
		string(scopesJSON),
		account.TokenType,
		nullTime(account.ExpiresAt),
		nullTime(account.LastValidatedAt),
	)
	if err != nil {
		return fmt.Errorf("save twitch account %q: %w", account.Kind, err)
	}

	return nil
}

func (s *TwitchAccountStore) Get(ctx context.Context, kind TwitchAccountKind) (*TwitchAccount, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	query := `
SELECT
	kind,
	twitch_user_id,
	login,
	display_name,
	access_token,
	refresh_token,
	scopes,
	token_type,
	expires_at,
	last_validated_at,
	created_at,
	updated_at
FROM twitch_accounts
WHERE kind = $1
`

	var (
		account       TwitchAccount
		scopesJSON    string
		expiresAt     sql.NullTime
		lastValidated sql.NullTime
	)

	err = db.QueryRowContext(ctx, query, kind).Scan(
		&account.Kind,
		&account.TwitchUserID,
		&account.Login,
		&account.DisplayName,
		&account.AccessToken,
		&account.RefreshToken,
		&scopesJSON,
		&account.TokenType,
		&expiresAt,
		&lastValidated,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get twitch account %q: %w", kind, err)
	}

	if scopesJSON != "" {
		if err := json.Unmarshal([]byte(scopesJSON), &account.Scopes); err != nil {
			return nil, fmt.Errorf("unmarshal twitch scopes: %w", err)
		}
	}
	if expiresAt.Valid {
		account.ExpiresAt = expiresAt.Time
	}
	if lastValidated.Valid {
		account.LastValidatedAt = lastValidated.Time
	}

	return &account, nil
}

func (s *TwitchAccountStore) Delete(ctx context.Context, kind TwitchAccountKind) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, `DELETE FROM twitch_accounts WHERE kind = $1`, kind); err != nil {
		return fmt.Errorf("delete twitch account %q: %w", kind, err)
	}

	return nil
}

func nullTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}

	return value
}
