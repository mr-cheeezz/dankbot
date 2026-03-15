package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type RobloxAccountKind string

const (
	RobloxAccountKindStreamer RobloxAccountKind = "streamer"
)

type RobloxAccount struct {
	Kind         RobloxAccountKind
	RobloxUserID string
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

type RobloxAccountStore struct {
	client *Client
}

func NewRobloxAccountStore(client *Client) *RobloxAccountStore {
	return &RobloxAccountStore{client: client}
}

func (s *RobloxAccountStore) Save(ctx context.Context, account RobloxAccount) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	if account.DisplayName == "" {
		account.DisplayName = account.Username
	}

	query := `
INSERT INTO roblox_accounts (
	kind,
	roblox_user_id,
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
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
ON CONFLICT (kind) DO UPDATE SET
	roblox_user_id = EXCLUDED.roblox_user_id,
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
		account.RobloxUserID,
		account.Username,
		account.DisplayName,
		account.AccessToken,
		account.RefreshToken,
		account.Scope,
		account.TokenType,
		nullTime(account.ExpiresAt),
	)
	if err != nil {
		return fmt.Errorf("save roblox account %q: %w", account.Kind, err)
	}

	return nil
}

func (s *RobloxAccountStore) Get(ctx context.Context, kind RobloxAccountKind) (*RobloxAccount, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	query := `
SELECT
	kind,
	roblox_user_id,
	username,
	display_name,
	access_token,
	refresh_token,
	scope,
	token_type,
	expires_at,
	created_at,
	updated_at
FROM roblox_accounts
WHERE kind = $1
`

	var (
		account   RobloxAccount
		expiresAt sql.NullTime
	)

	err = db.QueryRowContext(ctx, query, kind).Scan(
		&account.Kind,
		&account.RobloxUserID,
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
		return nil, fmt.Errorf("get roblox account %q: %w", kind, err)
	}

	if expiresAt.Valid {
		account.ExpiresAt = expiresAt.Time
	}

	return &account, nil
}

func (s *RobloxAccountStore) Delete(ctx context.Context, kind RobloxAccountKind) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, `DELETE FROM roblox_accounts WHERE kind = $1`, kind); err != nil {
		return fmt.Errorf("delete roblox account %q: %w", kind, err)
	}

	return nil
}
