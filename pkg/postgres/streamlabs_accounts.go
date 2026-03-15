package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type StreamlabsAccountKind string

const (
	StreamlabsAccountKindStreamer StreamlabsAccountKind = "streamer"
)

type StreamlabsAccount struct {
	Kind             StreamlabsAccountKind
	StreamlabsUserID string
	DisplayName      string
	AccessToken      string
	RefreshToken     string
	Scope            string
	TokenType        string
	ExpiresAt        time.Time
	SocketToken      string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type StreamlabsAccountStore struct {
	client *Client
}

func NewStreamlabsAccountStore(client *Client) *StreamlabsAccountStore {
	return &StreamlabsAccountStore{client: client}
}

func (s *StreamlabsAccountStore) Save(ctx context.Context, account StreamlabsAccount) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	query := `
INSERT INTO streamlabs_accounts (
	kind,
	streamlabs_user_id,
	display_name,
	access_token,
	refresh_token,
	scope,
	token_type,
	expires_at,
	socket_token,
	created_at,
	updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
ON CONFLICT (kind) DO UPDATE SET
	streamlabs_user_id = EXCLUDED.streamlabs_user_id,
	display_name = EXCLUDED.display_name,
	access_token = EXCLUDED.access_token,
	refresh_token = EXCLUDED.refresh_token,
	scope = EXCLUDED.scope,
	token_type = EXCLUDED.token_type,
	expires_at = EXCLUDED.expires_at,
	socket_token = EXCLUDED.socket_token,
	updated_at = NOW()
`

	_, err = db.ExecContext(
		ctx,
		query,
		account.Kind,
		account.StreamlabsUserID,
		account.DisplayName,
		account.AccessToken,
		account.RefreshToken,
		account.Scope,
		account.TokenType,
		nullTime(account.ExpiresAt),
		account.SocketToken,
	)
	if err != nil {
		return fmt.Errorf("save streamlabs account %q: %w", account.Kind, err)
	}

	return nil
}

func (s *StreamlabsAccountStore) Get(ctx context.Context, kind StreamlabsAccountKind) (*StreamlabsAccount, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	query := `
SELECT
	kind,
	streamlabs_user_id,
	display_name,
	access_token,
	refresh_token,
	scope,
	token_type,
	expires_at,
	socket_token,
	created_at,
	updated_at
FROM streamlabs_accounts
WHERE kind = $1
`

	var (
		account   StreamlabsAccount
		expiresAt sql.NullTime
	)

	err = db.QueryRowContext(ctx, query, kind).Scan(
		&account.Kind,
		&account.StreamlabsUserID,
		&account.DisplayName,
		&account.AccessToken,
		&account.RefreshToken,
		&account.Scope,
		&account.TokenType,
		&expiresAt,
		&account.SocketToken,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get streamlabs account %q: %w", kind, err)
	}

	if expiresAt.Valid {
		account.ExpiresAt = expiresAt.Time
	}

	return &account, nil
}

func (s *StreamlabsAccountStore) Delete(ctx context.Context, kind StreamlabsAccountKind) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, `DELETE FROM streamlabs_accounts WHERE kind = $1`, kind); err != nil {
		return fmt.Errorf("delete streamlabs account %q: %w", kind, err)
	}

	return nil
}
