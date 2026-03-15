package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type BlockedTerm struct {
	ID             string
	Pattern        string
	IsRegex        bool
	Action         string
	TimeoutSeconds int
	Reason         string
	Enabled        bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type BlockedTermStore struct {
	client *Client
}

func NewBlockedTermStore(client *Client) *BlockedTermStore {
	return &BlockedTermStore{client: client}
}

func (s *BlockedTermStore) List(ctx context.Context) ([]BlockedTerm, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(
		ctx,
		`
SELECT
	id,
	pattern,
	is_regex,
	action,
	timeout_seconds,
	reason,
	enabled,
	created_at,
	updated_at
FROM blocked_terms
ORDER BY enabled DESC, pattern ASC, created_at ASC
`,
	)
	if err != nil {
		return nil, fmt.Errorf("list blocked terms: %w", err)
	}
	defer rows.Close()

	items := make([]BlockedTerm, 0)
	for rows.Next() {
		var item BlockedTerm
		if err := rows.Scan(
			&item.ID,
			&item.Pattern,
			&item.IsRegex,
			&item.Action,
			&item.TimeoutSeconds,
			&item.Reason,
			&item.Enabled,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan blocked term: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate blocked terms: %w", err)
	}

	return items, nil
}

func (s *BlockedTermStore) Get(ctx context.Context, id string) (*BlockedTerm, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	id = strings.TrimSpace(id)
	if id == "" {
		return nil, nil
	}

	var item BlockedTerm
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	id,
	pattern,
	is_regex,
	action,
	timeout_seconds,
	reason,
	enabled,
	created_at,
	updated_at
FROM blocked_terms
WHERE id = $1
`,
		id,
	).Scan(
		&item.ID,
		&item.Pattern,
		&item.IsRegex,
		&item.Action,
		&item.TimeoutSeconds,
		&item.Reason,
		&item.Enabled,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get blocked term %q: %w", id, err)
	}

	return &item, nil
}

func (s *BlockedTermStore) Create(ctx context.Context, term BlockedTerm) (*BlockedTerm, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	term = normalizeBlockedTerm(term)
	if err := validateBlockedTerm(term); err != nil {
		return nil, err
	}

	_, err = db.ExecContext(
		ctx,
		`
INSERT INTO blocked_terms (
	id,
	pattern,
	is_regex,
	action,
	timeout_seconds,
	reason,
	enabled,
	created_at,
	updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
`,
		term.ID,
		term.Pattern,
		term.IsRegex,
		term.Action,
		term.TimeoutSeconds,
		term.Reason,
		term.Enabled,
	)
	if err != nil {
		return nil, fmt.Errorf("create blocked term %q: %w", term.ID, err)
	}

	return s.Get(ctx, term.ID)
}

func (s *BlockedTermStore) Update(ctx context.Context, term BlockedTerm) (*BlockedTerm, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	term = normalizeBlockedTerm(term)
	if err := validateBlockedTerm(term); err != nil {
		return nil, err
	}

	result, err := db.ExecContext(
		ctx,
		`
UPDATE blocked_terms
SET
	pattern = $2,
	is_regex = $3,
	action = $4,
	timeout_seconds = $5,
	reason = $6,
	enabled = $7,
	updated_at = NOW()
WHERE id = $1
`,
		term.ID,
		term.Pattern,
		term.IsRegex,
		term.Action,
		term.TimeoutSeconds,
		term.Reason,
		term.Enabled,
	)
	if err != nil {
		return nil, fmt.Errorf("update blocked term %q: %w", term.ID, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("blocked term %q rows affected: %w", term.ID, err)
	}
	if rowsAffected == 0 {
		return nil, fmt.Errorf("blocked term %q does not exist", term.ID)
	}

	return s.Get(ctx, term.ID)
}

func (s *BlockedTermStore) Delete(ctx context.Context, id string) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	id = strings.TrimSpace(id)
	if id == "" {
		return nil
	}

	if _, err := db.ExecContext(ctx, `DELETE FROM blocked_terms WHERE id = $1`, id); err != nil {
		return fmt.Errorf("delete blocked term %q: %w", id, err)
	}

	return nil
}

func normalizeBlockedTerm(term BlockedTerm) BlockedTerm {
	term.ID = strings.TrimSpace(term.ID)
	term.Pattern = strings.TrimSpace(term.Pattern)
	term.Action = NormalizeBlockedTermActionForAPI(term.Action)
	term.Reason = strings.TrimSpace(term.Reason)
	if term.TimeoutSeconds < 0 {
		term.TimeoutSeconds = 0
	}
	return term
}

func validateBlockedTerm(term BlockedTerm) error {
	if term.ID == "" {
		return fmt.Errorf("blocked term id is required")
	}
	if term.Pattern == "" {
		return fmt.Errorf("blocked term pattern is required")
	}
	if term.Action == "" {
		return fmt.Errorf("blocked term action is required")
	}
	if term.Action == "timeout" || term.Action == "delete + timeout" {
		if term.TimeoutSeconds <= 0 {
			return fmt.Errorf("blocked term timeout must be greater than zero")
		}
	}

	return nil
}

func NormalizeBlockedTermActionForAPI(action string) string {
	switch strings.TrimSpace(strings.ToLower(action)) {
	case "delete":
		return "delete"
	case "warn":
		return "warn"
	case "delete + warn", "delete+warn":
		return "delete + warn"
	case "timeout":
		return "timeout"
	case "delete + timeout", "delete+timeout":
		return "delete + timeout"
	case "ban":
		return "ban"
	case "delete + ban", "delete+ban":
		return "delete + ban"
	default:
		return ""
	}
}
