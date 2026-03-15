package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Keyword struct {
	ID        int64
	Trigger   string
	Response  string
	CreatedBy string
	UpdatedBy string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type KeywordStore struct {
	client *Client
}

func NewKeywordStore(client *Client) *KeywordStore {
	return &KeywordStore{client: client}
}

func (s *KeywordStore) List(ctx context.Context) ([]Keyword, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(
		ctx,
		`
SELECT
	id,
	trigger,
	response,
	created_by,
	updated_by,
	created_at,
	updated_at
FROM keywords
ORDER BY length(trigger) DESC, trigger ASC
`,
	)
	if err != nil {
		return nil, fmt.Errorf("list keywords: %w", err)
	}
	defer rows.Close()

	keywords := make([]Keyword, 0)
	for rows.Next() {
		var keyword Keyword
		if err := rows.Scan(
			&keyword.ID,
			&keyword.Trigger,
			&keyword.Response,
			&keyword.CreatedBy,
			&keyword.UpdatedBy,
			&keyword.CreatedAt,
			&keyword.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan keyword: %w", err)
		}
		keywords = append(keywords, keyword)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate keywords: %w", err)
	}

	return keywords, nil
}

func (s *KeywordStore) GetByTrigger(ctx context.Context, trigger string) (*Keyword, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	trigger = normalizeKeywordTrigger(trigger)
	if trigger == "" {
		return nil, nil
	}

	var keyword Keyword
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	id,
	trigger,
	response,
	created_by,
	updated_by,
	created_at,
	updated_at
FROM keywords
WHERE LOWER(trigger) = LOWER($1)
`,
		trigger,
	).Scan(
		&keyword.ID,
		&keyword.Trigger,
		&keyword.Response,
		&keyword.CreatedBy,
		&keyword.UpdatedBy,
		&keyword.CreatedAt,
		&keyword.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get keyword %q: %w", trigger, err)
	}

	return &keyword, nil
}

func (s *KeywordStore) Create(ctx context.Context, trigger, response, createdBy string) (*Keyword, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	trigger = normalizeKeywordTrigger(trigger)
	response = strings.TrimSpace(response)
	if trigger == "" {
		return nil, fmt.Errorf("keyword trigger is required")
	}
	if response == "" {
		return nil, fmt.Errorf("keyword response is required")
	}

	var keyword Keyword
	err = db.QueryRowContext(
		ctx,
		`
INSERT INTO keywords (
	trigger,
	response,
	created_by,
	updated_by,
	created_at,
	updated_at
)
VALUES ($1, $2, $3, $3, NOW(), NOW())
RETURNING id, trigger, response, created_by, updated_by, created_at, updated_at
`,
		trigger,
		response,
		strings.TrimSpace(createdBy),
	).Scan(
		&keyword.ID,
		&keyword.Trigger,
		&keyword.Response,
		&keyword.CreatedBy,
		&keyword.UpdatedBy,
		&keyword.CreatedAt,
		&keyword.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create keyword %q: %w", trigger, err)
	}

	return &keyword, nil
}

func (s *KeywordStore) UpdateByTrigger(ctx context.Context, trigger, response, updatedBy string) (*Keyword, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	trigger = normalizeKeywordTrigger(trigger)
	response = strings.TrimSpace(response)
	if trigger == "" {
		return nil, fmt.Errorf("keyword trigger is required")
	}
	if response == "" {
		return nil, fmt.Errorf("keyword response is required")
	}

	var keyword Keyword
	err = db.QueryRowContext(
		ctx,
		`
UPDATE keywords
SET
	response = $2,
	updated_by = $3,
	updated_at = NOW()
WHERE LOWER(trigger) = LOWER($1)
RETURNING id, trigger, response, created_by, updated_by, created_at, updated_at
`,
		trigger,
		response,
		strings.TrimSpace(updatedBy),
	).Scan(
		&keyword.ID,
		&keyword.Trigger,
		&keyword.Response,
		&keyword.CreatedBy,
		&keyword.UpdatedBy,
		&keyword.CreatedAt,
		&keyword.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("update keyword %q: %w", trigger, err)
	}

	return &keyword, nil
}

func (s *KeywordStore) DeleteByTrigger(ctx context.Context, trigger string) (bool, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return false, err
	}

	trigger = normalizeKeywordTrigger(trigger)
	if trigger == "" {
		return false, fmt.Errorf("keyword trigger is required")
	}

	result, err := db.ExecContext(ctx, `DELETE FROM keywords WHERE LOWER(trigger) = LOWER($1)`, trigger)
	if err != nil {
		return false, fmt.Errorf("delete keyword %q: %w", trigger, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("delete keyword %q rows affected: %w", trigger, err)
	}

	return rows > 0, nil
}

func normalizeKeywordTrigger(trigger string) string {
	return strings.TrimSpace(trigger)
}
