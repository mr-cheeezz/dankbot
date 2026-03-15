package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type SpotifyAccountKind string

const (
	SpotifyAccountKindStreamer SpotifyAccountKind = "streamer"
)

type SpotifyAccount struct {
	Kind          SpotifyAccountKind
	SpotifyUserID string
	DisplayName   string
	Email         string
	Product       string
	Country       string
	AccessToken   string
	RefreshToken  string
	Scope         string
	TokenType     string
	ExpiresAt     time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type SpotifyAccountStore struct {
	client *Client
}

func NewSpotifyAccountStore(client *Client) *SpotifyAccountStore {
	return &SpotifyAccountStore{client: client}
}

func (s *SpotifyAccountStore) Save(ctx context.Context, account SpotifyAccount) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	query := `
INSERT INTO spotify_accounts (
	kind,
	spotify_user_id,
	display_name,
	email,
	product,
	country,
	access_token,
	refresh_token,
	scope,
	token_type,
	expires_at,
	created_at,
	updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
ON CONFLICT (kind) DO UPDATE SET
	spotify_user_id = EXCLUDED.spotify_user_id,
	display_name = EXCLUDED.display_name,
	email = EXCLUDED.email,
	product = EXCLUDED.product,
	country = EXCLUDED.country,
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
		account.SpotifyUserID,
		account.DisplayName,
		account.Email,
		account.Product,
		account.Country,
		account.AccessToken,
		account.RefreshToken,
		account.Scope,
		account.TokenType,
		nullTime(account.ExpiresAt),
	)
	if err != nil {
		return fmt.Errorf("save spotify account %q: %w", account.Kind, err)
	}

	return nil
}

func (s *SpotifyAccountStore) Get(ctx context.Context, kind SpotifyAccountKind) (*SpotifyAccount, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	query := `
SELECT
	kind,
	spotify_user_id,
	display_name,
	email,
	product,
	country,
	access_token,
	refresh_token,
	scope,
	token_type,
	expires_at,
	created_at,
	updated_at
FROM spotify_accounts
WHERE kind = $1
`

	var (
		account   SpotifyAccount
		expiresAt sql.NullTime
	)

	err = db.QueryRowContext(ctx, query, kind).Scan(
		&account.Kind,
		&account.SpotifyUserID,
		&account.DisplayName,
		&account.Email,
		&account.Product,
		&account.Country,
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
		return nil, fmt.Errorf("get spotify account %q: %w", kind, err)
	}

	if expiresAt.Valid {
		account.ExpiresAt = expiresAt.Time
	}

	return &account, nil
}

func (s *SpotifyAccountStore) Delete(ctx context.Context, kind SpotifyAccountKind) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, `DELETE FROM spotify_accounts WHERE kind = $1`, kind); err != nil {
		return fmt.Errorf("delete spotify account %q: %w", kind, err)
	}

	return nil
}
