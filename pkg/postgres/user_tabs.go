package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"
)

type UserTab struct {
	Login          string
	DisplayName    string
	BalanceCents   int64
	LastInterestAt time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type UserTabEvent struct {
	ID           int64
	Login        string
	Action       string
	AmountCents  int64
	BalanceCents int64
	Note         string
	CreatedAt    time.Time
}

type UserTabStore struct {
	client *Client
}

func NewUserTabStore(client *Client) *UserTabStore {
	return &UserTabStore{client: client}
}

func (s *UserTabStore) Get(ctx context.Context, login string) (*UserTab, int64, error) {
	return s.getWithInterest(ctx, login, 0, 0)
}

func (s *UserTabStore) Add(ctx context.Context, login, displayName string, deltaCents int64, interestRatePct float64, interestEveryDays int) (*UserTab, int64, error) {
	return s.mutate(ctx, login, displayName, interestRatePct, interestEveryDays, "add", deltaCents, "", func(current int64) int64 {
		return current + deltaCents
	})
}

func (s *UserTabStore) Set(ctx context.Context, login, displayName string, nextCents int64, interestRatePct float64, interestEveryDays int) (*UserTab, int64, error) {
	return s.mutate(ctx, login, displayName, interestRatePct, interestEveryDays, "set", nextCents, "", func(current int64) int64 {
		_ = current
		return nextCents
	})
}

func (s *UserTabStore) MarkPaid(ctx context.Context, login, displayName string, interestRatePct float64, interestEveryDays int) (*UserTab, int64, error) {
	return s.mutate(ctx, login, displayName, interestRatePct, interestEveryDays, "paid", 0, "paid off", func(current int64) int64 {
		_ = current
		return 0
	})
}

func (s *UserTabStore) Ensure(ctx context.Context, login, displayName string, interestRatePct float64, interestEveryDays int) (*UserTab, int64, error) {
	return s.mutate(ctx, login, displayName, interestRatePct, interestEveryDays, "give", 0, "opened tab", func(current int64) int64 {
		return current
	})
}

func (s *UserTabStore) ListEvents(ctx context.Context, login string, limit int, offset int) ([]UserTabEvent, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	login = normalizeTabLogin(login)
	if login == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 5
	}
	if limit > 250 {
		limit = 250
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := db.QueryContext(
		ctx,
		`
SELECT id, login, action, amount_cents, balance_cents, note, created_at
FROM user_tab_events
WHERE login = $1
ORDER BY created_at DESC, id DESC
LIMIT $2 OFFSET $3
`,
		login,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list user tab events %q: %w", login, err)
	}
	defer rows.Close()

	items := make([]UserTabEvent, 0, limit)
	for rows.Next() {
		var entry UserTabEvent
		if err := rows.Scan(
			&entry.ID,
			&entry.Login,
			&entry.Action,
			&entry.AmountCents,
			&entry.BalanceCents,
			&entry.Note,
			&entry.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user tab event %q: %w", login, err)
		}
		items = append(items, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user tab events %q: %w", login, err)
	}

	return items, nil
}

func (s *UserTabStore) getWithInterest(ctx context.Context, login string, interestRatePct float64, interestEveryDays int) (*UserTab, int64, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, 0, err
	}

	login = normalizeTabLogin(login)
	if login == "" {
		return nil, 0, nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("begin user tab transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	row, err := getTabForUpdate(ctx, tx, login)
	if err != nil {
		return nil, 0, err
	}
	if row == nil {
		if err = tx.Commit(); err != nil {
			return nil, 0, fmt.Errorf("commit user tab read transaction: %w", err)
		}
		return nil, 0, nil
	}

	interestCents, changed := applyInterest(row, normalizeTabsInterestRate(interestRatePct), normalizeTabsInterestEveryDays(interestEveryDays), time.Now().UTC())
	if changed {
		if err := updateTabFields(ctx, tx, row); err != nil {
			return nil, 0, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, 0, fmt.Errorf("commit user tab transaction: %w", err)
	}

	return row, interestCents, nil
}

func (s *UserTabStore) mutate(
	ctx context.Context,
	login string,
	displayName string,
	interestRatePct float64,
	interestEveryDays int,
	action string,
	amountCents int64,
	note string,
	nextBalance func(current int64) int64,
) (*UserTab, int64, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, 0, err
	}

	login = normalizeTabLogin(login)
	displayName = normalizeTabDisplayName(displayName, login)
	if login == "" {
		return nil, 0, nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("begin user tab mutation transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	row, err := getTabForUpdate(ctx, tx, login)
	if err != nil {
		return nil, 0, err
	}
	now := time.Now().UTC()
	if row == nil {
		row = &UserTab{
			Login:          login,
			DisplayName:    displayName,
			BalanceCents:   0,
			LastInterestAt: now,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
	}

	interestCents, _ := applyInterest(row, normalizeTabsInterestRate(interestRatePct), normalizeTabsInterestEveryDays(interestEveryDays), now)
	row.BalanceCents = nextBalance(row.BalanceCents)
	if row.BalanceCents < 0 {
		row.BalanceCents = 0
	}
	if displayName != "" {
		row.DisplayName = displayName
	}
	row.UpdatedAt = now
	if row.BalanceCents == 0 {
		row.LastInterestAt = now
	}

	if err := upsertTab(ctx, tx, row); err != nil {
		return nil, 0, err
	}
	if err := insertTabEvent(ctx, tx, UserTabEvent{
		Login:        row.Login,
		Action:       strings.TrimSpace(action),
		AmountCents:  amountCents,
		BalanceCents: row.BalanceCents,
		Note:         strings.TrimSpace(note),
		CreatedAt:    now,
	}); err != nil {
		return nil, 0, err
	}
	if err = tx.Commit(); err != nil {
		return nil, 0, fmt.Errorf("commit user tab mutation transaction: %w", err)
	}

	return row, interestCents, nil
}

func getTabForUpdate(ctx context.Context, tx *sql.Tx, login string) (*UserTab, error) {
	var row UserTab
	err := tx.QueryRowContext(
		ctx,
		`
SELECT
	login,
	display_name,
	balance_cents,
	last_interest_at,
	created_at,
	updated_at
FROM user_tabs
WHERE login = $1
FOR UPDATE
`,
		login,
	).Scan(
		&row.Login,
		&row.DisplayName,
		&row.BalanceCents,
		&row.LastInterestAt,
		&row.CreatedAt,
		&row.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user tab %q: %w", login, err)
	}

	return &row, nil
}

func applyInterest(row *UserTab, interestRatePct float64, interestEveryDays int, now time.Time) (int64, bool) {
	if row == nil || row.BalanceCents <= 0 || interestRatePct <= 0 || interestEveryDays <= 0 {
		return 0, false
	}

	base := row.LastInterestAt
	if base.IsZero() {
		base = row.UpdatedAt
	}
	if base.IsZero() {
		base = row.CreatedAt
	}
	if base.IsZero() {
		base = now
	}
	if now.Before(base) {
		return 0, false
	}

	interval := time.Duration(interestEveryDays) * 24 * time.Hour
	if interval <= 0 {
		return 0, false
	}

	periods := int(now.Sub(base) / interval)
	if periods < 1 {
		return 0, false
	}

	rateFactor := 1 + (interestRatePct / 100.0)
	newBalance := row.BalanceCents
	for i := 0; i < periods; i++ {
		newBalance = int64(math.Round(float64(newBalance) * rateFactor))
	}
	if newBalance < row.BalanceCents {
		newBalance = row.BalanceCents
	}

	interestCents := newBalance - row.BalanceCents
	row.BalanceCents = newBalance
	row.LastInterestAt = base.Add(time.Duration(periods) * interval)
	row.UpdatedAt = now
	return interestCents, true
}

func updateTabFields(ctx context.Context, tx *sql.Tx, row *UserTab) error {
	if row == nil {
		return nil
	}
	_, err := tx.ExecContext(
		ctx,
		`
UPDATE user_tabs
SET
	display_name = $2,
	balance_cents = $3,
	last_interest_at = $4,
	updated_at = $5
WHERE login = $1
`,
		row.Login,
		row.DisplayName,
		row.BalanceCents,
		row.LastInterestAt.UTC(),
		row.UpdatedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("update user tab %q: %w", row.Login, err)
	}
	return nil
}

func upsertTab(ctx context.Context, tx *sql.Tx, row *UserTab) error {
	if row == nil {
		return nil
	}
	_, err := tx.ExecContext(
		ctx,
		`
INSERT INTO user_tabs (
	login,
	display_name,
	balance_cents,
	last_interest_at,
	created_at,
	updated_at
)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (login) DO UPDATE SET
	display_name = EXCLUDED.display_name,
	balance_cents = EXCLUDED.balance_cents,
	last_interest_at = EXCLUDED.last_interest_at,
	updated_at = EXCLUDED.updated_at
`,
		row.Login,
		row.DisplayName,
		row.BalanceCents,
		row.LastInterestAt.UTC(),
		row.CreatedAt.UTC(),
		row.UpdatedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("upsert user tab %q: %w", row.Login, err)
	}
	return nil
}

func normalizeTabLogin(login string) string {
	login = strings.TrimSpace(strings.ToLower(strings.TrimPrefix(login, "@")))
	if login == "" {
		return ""
	}
	return login
}

func normalizeTabDisplayName(displayName, login string) string {
	displayName = strings.TrimSpace(displayName)
	if displayName != "" {
		return displayName
	}
	return normalizeTabLogin(login)
}

func insertTabEvent(ctx context.Context, tx *sql.Tx, event UserTabEvent) error {
	action := strings.TrimSpace(event.Action)
	if action == "" {
		action = "set"
	}
	createdAt := event.CreatedAt.UTC()
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	_, err := tx.ExecContext(
		ctx,
		`
INSERT INTO user_tab_events (
	login,
	action,
	amount_cents,
	balance_cents,
	note,
	created_at
)
VALUES ($1, $2, $3, $4, $5, $6)
`,
		event.Login,
		action,
		event.AmountCents,
		event.BalanceCents,
		strings.TrimSpace(event.Note),
		createdAt,
	)
	if err != nil {
		return fmt.Errorf("insert user tab event %q: %w", event.Login, err)
	}
	return nil
}
