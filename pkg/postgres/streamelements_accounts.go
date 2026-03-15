package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type StreamElementsAccountKind string

const (
	StreamElementsAccountKindStreamer StreamElementsAccountKind = "streamer"
)

type StreamElementsAccount struct {
	Kind         StreamElementsAccountKind
	ChannelID    string
	Provider     string
	Username     string
	DisplayName  string
	AccessToken  string
	RefreshToken string
	Scope        string
	TokenType    string
	ExpiresAt    time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type StreamElementsAccountStore struct {
	client *Client
}

func NewStreamElementsAccountStore(client *Client) *StreamElementsAccountStore {
	return &StreamElementsAccountStore{client: client}
}

func (s *StreamElementsAccountStore) Save(ctx context.Context, account StreamElementsAccount) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	query := `
INSERT INTO streamelements_accounts (
	kind,
	channel_id,
	provider,
	username,
	display_name,
	access_token,
	refresh_token,
	scope,
	token_type,
	expires_at,
	created_at,
	updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
ON CONFLICT (kind) DO UPDATE SET
	channel_id = EXCLUDED.channel_id,
	provider = EXCLUDED.provider,
	username = EXCLUDED.username,
	display_name = EXCLUDED.display_name,
	access_token = EXCLUDED.access_token,
	refresh_token = EXCLUDED.refresh_token,
	scope = EXCLUDED.scope,
	token_type = EXCLUDED.token_type,
	expires_at = EXCLUDED.expires_at,
	updated_at = NOW()
`

	_, err = db.ExecContext(
		ctx,
		query,
		account.Kind,
		account.ChannelID,
		account.Provider,
		account.Username,
		account.DisplayName,
		account.AccessToken,
		account.RefreshToken,
		account.Scope,
		account.TokenType,
		nullTime(account.ExpiresAt),
	)
	if err != nil {
		return fmt.Errorf("save streamelements account %q: %w", account.Kind, err)
	}

	return nil
}

func (s *StreamElementsAccountStore) Get(ctx context.Context, kind StreamElementsAccountKind) (*StreamElementsAccount, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	query := `
SELECT
	kind,
	channel_id,
	provider,
	username,
	display_name,
	access_token,
	refresh_token,
	scope,
	token_type,
	expires_at,
	created_at,
	updated_at
FROM streamelements_accounts
WHERE kind = $1
`

	var (
		account   StreamElementsAccount
		expiresAt sql.NullTime
	)

	err = db.QueryRowContext(ctx, query, kind).Scan(
		&account.Kind,
		&account.ChannelID,
		&account.Provider,
		&account.Username,
		&account.DisplayName,
		&account.AccessToken,
		&account.RefreshToken,
		&account.Scope,
		&account.TokenType,
		&expiresAt,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get streamelements account %q: %w", kind, err)
	}

	if expiresAt.Valid {
		account.ExpiresAt = expiresAt.Time
	}

	return &account, nil
}

func (s *StreamElementsAccountStore) Delete(ctx context.Context, kind StreamElementsAccountKind) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, `DELETE FROM streamelements_accounts WHERE kind = $1`, kind); err != nil {
		return fmt.Errorf("delete streamelements account %q: %w", kind, err)
	}

	return nil
}
