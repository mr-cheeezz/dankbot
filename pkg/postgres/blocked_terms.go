package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type BlockedTerm struct {
	ID             string
	Name           string
	Pattern        string
	PhraseGroups   [][]string
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
	name,
	pattern,
	phrase_groups,
	is_regex,
	action,
	timeout_seconds,
	reason,
	enabled,
	created_at,
	updated_at
FROM blocked_terms
ORDER BY enabled DESC, name ASC, created_at ASC
`,
	)
	if err != nil {
		return nil, fmt.Errorf("list blocked terms: %w", err)
	}
	defer rows.Close()

	items := make([]BlockedTerm, 0)
	for rows.Next() {
		var item BlockedTerm
		var phraseGroupsRaw []byte
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Pattern,
			&phraseGroupsRaw,
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
		if len(phraseGroupsRaw) > 0 {
			if err := json.Unmarshal(phraseGroupsRaw, &item.PhraseGroups); err != nil {
				return nil, fmt.Errorf("unmarshal blocked term phrase groups: %w", err)
			}
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
	var phraseGroupsRaw []byte
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	id,
	name,
	pattern,
	phrase_groups,
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
		&item.Name,
		&item.Pattern,
		&phraseGroupsRaw,
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
	if len(phraseGroupsRaw) > 0 {
		if err := json.Unmarshal(phraseGroupsRaw, &item.PhraseGroups); err != nil {
			return nil, fmt.Errorf("unmarshal blocked term phrase groups: %w", err)
		}
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
	phraseGroupsRaw, err := json.Marshal(term.PhraseGroups)
	if err != nil {
		return nil, fmt.Errorf("marshal blocked term phrase groups: %w", err)
	}

	_, err = db.ExecContext(
		ctx,
		`
INSERT INTO blocked_terms (
	id,
	name,
	pattern,
	phrase_groups,
	is_regex,
	action,
	timeout_seconds,
	reason,
	enabled,
	created_at,
	updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
`,
		term.ID,
		term.Name,
		term.Pattern,
		phraseGroupsRaw,
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
	phraseGroupsRaw, err := json.Marshal(term.PhraseGroups)
	if err != nil {
		return nil, fmt.Errorf("marshal blocked term phrase groups: %w", err)
	}

	result, err := db.ExecContext(
		ctx,
		`
UPDATE blocked_terms
SET
	name = $2,
	pattern = $3,
	phrase_groups = $4,
	is_regex = $5,
	action = $6,
	timeout_seconds = $7,
	reason = $8,
	enabled = $9,
	updated_at = NOW()
WHERE id = $1
`,
		term.ID,
		term.Name,
		term.Pattern,
		phraseGroupsRaw,
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
	term.Name = strings.TrimSpace(term.Name)
	term.Pattern = strings.TrimSpace(term.Pattern)
	term.PhraseGroups = normalizeBlockedTermPhraseGroups(term.PhraseGroups)
	term.Action = NormalizeBlockedTermActionForAPI(term.Action)
	term.Reason = strings.TrimSpace(term.Reason)
	if !term.IsRegex && len(term.PhraseGroups) == 0 && term.Pattern != "" {
		term.PhraseGroups = [][]string{{term.Pattern}}
	}
	if term.Name == "" {
		switch {
		case term.Pattern != "":
			term.Name = term.Pattern
		case len(term.PhraseGroups) > 0 && len(term.PhraseGroups[0]) > 0:
			term.Name = strings.Join(term.PhraseGroups[0], " + ")
		}
	}
	if term.TimeoutSeconds < 0 {
		term.TimeoutSeconds = 0
	}
	return term
}

func validateBlockedTerm(term BlockedTerm) error {
	if term.ID == "" {
		return fmt.Errorf("blocked term id is required")
	}
	if term.Name == "" {
		return fmt.Errorf("blocked term name is required")
	}
	if term.IsRegex {
		if term.Pattern == "" {
			return fmt.Errorf("blocked term pattern is required")
		}
	} else if term.Pattern == "" && len(term.PhraseGroups) == 0 {
		return fmt.Errorf("blocked term needs at least one phrase group")
	}
	if term.Action == "" {
		return fmt.Errorf("blocked term action is required")
	}
	if term.Action == "timeout" {
		if term.TimeoutSeconds <= 0 {
			return fmt.Errorf("blocked term timeout must be greater than zero")
		}
	}

	return nil
}

func normalizeBlockedTermPhraseGroups(groups [][]string) [][]string {
	normalized := make([][]string, 0, len(groups))

	for _, group := range groups {
		nextGroup := make([]string, 0, len(group))
		for _, phrase := range group {
			value := strings.TrimSpace(phrase)
			if value == "" {
				continue
			}
			nextGroup = append(nextGroup, value)
		}
		if len(nextGroup) == 0 {
			continue
		}
		normalized = append(normalized, nextGroup)
	}

	return normalized
}

func NormalizeBlockedTermActionForAPI(action string) string {
	switch strings.TrimSpace(strings.ToLower(action)) {
	case "delete":
		return "delete"
	case "warn", "delete + warn", "delete+warn", "warn + delete", "warn+delete":
		return "delete + warn"
	case "timeout", "delete + timeout", "delete+timeout":
		return "timeout"
	case "ban", "delete + ban", "delete+ban":
		return "ban"
	default:
		return ""
	}
}
