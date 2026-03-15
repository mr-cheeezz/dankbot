package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Quote struct {
	ID        int64
	Message   string
	CreatedBy string
	UpdatedBy string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type QuoteStore struct {
	client *Client
}

func NewQuoteStore(client *Client) *QuoteStore {
	return &QuoteStore{client: client}
}

func (s *QuoteStore) Create(ctx context.Context, message, createdBy string) (*Quote, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	message = strings.TrimSpace(message)
	if message == "" {
		return nil, fmt.Errorf("quote message is required")
	}

	var quote Quote
	err = db.QueryRowContext(
		ctx,
		`
INSERT INTO quotes (
	message,
	created_by,
	updated_by,
	created_at,
	updated_at
)
VALUES ($1, $2, $2, NOW(), NOW())
RETURNING id, message, created_by, updated_by, created_at, updated_at
`,
		message,
		strings.TrimSpace(createdBy),
	).Scan(
		&quote.ID,
		&quote.Message,
		&quote.CreatedBy,
		&quote.UpdatedBy,
		&quote.CreatedAt,
		&quote.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create quote: %w", err)
	}

	return &quote, nil
}

func (s *QuoteStore) Get(ctx context.Context, id int64) (*Quote, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	if id <= 0 {
		return nil, nil
	}

	var quote Quote
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	id,
	message,
	created_by,
	updated_by,
	created_at,
	updated_at
FROM quotes
WHERE id = $1
`,
		id,
	).Scan(
		&quote.ID,
		&quote.Message,
		&quote.CreatedBy,
		&quote.UpdatedBy,
		&quote.CreatedAt,
		&quote.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get quote %d: %w", id, err)
	}

	return &quote, nil
}

func (s *QuoteStore) Random(ctx context.Context) (*Quote, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var quote Quote
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	id,
	message,
	created_by,
	updated_by,
	created_at,
	updated_at
FROM quotes
ORDER BY random()
LIMIT 1
`,
	).Scan(
		&quote.ID,
		&quote.Message,
		&quote.CreatedBy,
		&quote.UpdatedBy,
		&quote.CreatedAt,
		&quote.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get random quote: %w", err)
	}

	return &quote, nil
}

func (s *QuoteStore) List(ctx context.Context, limit int) ([]Quote, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var (
		rows *sql.Rows
	)
	if limit > 0 {
		rows, err = db.QueryContext(
			ctx,
			`
SELECT
	id,
	message,
	created_by,
	updated_by,
	created_at,
	updated_at
FROM quotes
ORDER BY id DESC
LIMIT $1
`,
			limit,
		)
	} else {
		rows, err = db.QueryContext(
			ctx,
			`
SELECT
	id,
	message,
	created_by,
	updated_by,
	created_at,
	updated_at
FROM quotes
ORDER BY id DESC
`,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("list quotes: %w", err)
	}
	defer rows.Close()

	quotes := make([]Quote, 0)
	for rows.Next() {
		var quote Quote
		if err := rows.Scan(
			&quote.ID,
			&quote.Message,
			&quote.CreatedBy,
			&quote.UpdatedBy,
			&quote.CreatedAt,
			&quote.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan quote: %w", err)
		}
		quotes = append(quotes, quote)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate quotes: %w", err)
	}

	return quotes, nil
}

func (s *QuoteStore) Update(ctx context.Context, id int64, message, updatedBy string) (*Quote, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	message = strings.TrimSpace(message)
	if id <= 0 {
		return nil, fmt.Errorf("quote id must be greater than 0")
	}
	if message == "" {
		return nil, fmt.Errorf("quote message is required")
	}

	var quote Quote
	err = db.QueryRowContext(
		ctx,
		`
UPDATE quotes
SET
	message = $2,
	updated_by = $3,
	updated_at = NOW()
WHERE id = $1
RETURNING id, message, created_by, updated_by, created_at, updated_at
`,
		id,
		message,
		strings.TrimSpace(updatedBy),
	).Scan(
		&quote.ID,
		&quote.Message,
		&quote.CreatedBy,
		&quote.UpdatedBy,
		&quote.CreatedAt,
		&quote.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("update quote %d: %w", id, err)
	}

	return &quote, nil
}

func (s *QuoteStore) Delete(ctx context.Context, id int64) (bool, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return false, err
	}

	if id <= 0 {
		return false, fmt.Errorf("quote id must be greater than 0")
	}

	result, err := db.ExecContext(ctx, `DELETE FROM quotes WHERE id = $1`, id)
	if err != nil {
		return false, fmt.Errorf("delete quote %d: %w", id, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("delete quote %d rows affected: %w", id, err)
	}

	return rows > 0, nil
}
