package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type RobloxGamePlaytime struct {
	UniverseID   int64
	RootPlaceID  int64
	GameName     string
	TotalSeconds int64
	LastSeenAt   time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RobloxPlaytimeStore struct {
	client *Client
}

func NewRobloxPlaytimeStore(client *Client) *RobloxPlaytimeStore {
	return &RobloxPlaytimeStore{client: client}
}

func (s *RobloxPlaytimeStore) AddDuration(ctx context.Context, streamSessionID string, universeID, rootPlaceID int64, gameName string, startedAt, endedAt time.Time) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	duration := endedAt.Sub(startedAt)
	seconds := int64(duration / time.Second)
	if seconds <= 0 {
		return nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin roblox playtime transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(
		ctx,
		`
INSERT INTO roblox_game_playtime (
	universe_id,
	root_place_id,
	game_name,
	total_seconds,
	last_seen_at,
	created_at,
	updated_at
)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
ON CONFLICT (universe_id) DO UPDATE SET
	root_place_id = EXCLUDED.root_place_id,
	game_name = CASE
		WHEN EXCLUDED.game_name <> '' THEN EXCLUDED.game_name
		ELSE roblox_game_playtime.game_name
	END,
	total_seconds = roblox_game_playtime.total_seconds + EXCLUDED.total_seconds,
	last_seen_at = EXCLUDED.last_seen_at,
	updated_at = NOW()
`,
		universeID,
		rootPlaceID,
		gameName,
		seconds,
		endedAt.UTC(),
	); err != nil {
		return fmt.Errorf("add roblox playtime for universe %d: %w", universeID, err)
	}

	if strings.TrimSpace(streamSessionID) != "" {
		if _, err = tx.ExecContext(
			ctx,
			`
INSERT INTO roblox_game_playtime_sessions (
	stream_session_id,
	universe_id,
	root_place_id,
	game_name,
	started_at,
	ended_at,
	duration_seconds,
	created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
`,
			strings.TrimSpace(streamSessionID),
			universeID,
			rootPlaceID,
			gameName,
			startedAt.UTC(),
			endedAt.UTC(),
			seconds,
		); err != nil {
			return fmt.Errorf("insert roblox playtime session for universe %d: %w", universeID, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit roblox playtime transaction: %w", err)
	}

	return nil
}

func (s *RobloxPlaytimeStore) Get(ctx context.Context, universeID int64) (*RobloxGamePlaytime, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var item RobloxGamePlaytime
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	universe_id,
	root_place_id,
	game_name,
	total_seconds,
	last_seen_at,
	created_at,
	updated_at
FROM roblox_game_playtime
WHERE universe_id = $1
`,
		universeID,
	).Scan(
		&item.UniverseID,
		&item.RootPlaceID,
		&item.GameName,
		&item.TotalSeconds,
		&item.LastSeenAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get roblox playtime for universe %d: %w", universeID, err)
	}

	return &item, nil
}

func (s *RobloxPlaytimeStore) ListTop(ctx context.Context, limit int) ([]RobloxGamePlaytime, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 5
	}

	rows, err := db.QueryContext(
		ctx,
		`
SELECT
	universe_id,
	root_place_id,
	game_name,
	total_seconds,
	last_seen_at,
	created_at,
	updated_at
FROM roblox_game_playtime
ORDER BY total_seconds DESC, updated_at DESC, universe_id ASC
LIMIT $1
`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list top roblox playtime: %w", err)
	}
	defer rows.Close()

	var items []RobloxGamePlaytime
	for rows.Next() {
		var item RobloxGamePlaytime
		if err := rows.Scan(
			&item.UniverseID,
			&item.RootPlaceID,
			&item.GameName,
			&item.TotalSeconds,
			&item.LastSeenAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan roblox playtime row: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate roblox playtime rows: %w", err)
	}

	return items, nil
}

func (s *RobloxPlaytimeStore) ListTopByRange(ctx context.Context, start, end time.Time, limit int) ([]RobloxGamePlaytime, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 5
	}

	rows, err := db.QueryContext(
		ctx,
		`
SELECT
	universe_id,
	MAX(root_place_id) AS root_place_id,
	MAX(game_name) AS game_name,
	SUM(duration_seconds) AS total_seconds,
	MAX(ended_at) AS last_seen_at,
	MIN(started_at) AS created_at,
	MAX(ended_at) AS updated_at
FROM roblox_game_playtime_sessions
WHERE ended_at >= $1 AND ended_at < $2
GROUP BY universe_id
ORDER BY total_seconds DESC, updated_at DESC, universe_id ASC
LIMIT $3
`,
		start.UTC(),
		end.UTC(),
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list roblox playtime by range: %w", err)
	}
	defer rows.Close()

	return scanPlaytimeRows(rows)
}

func (s *RobloxPlaytimeStore) ListTopByStreamSession(ctx context.Context, streamSessionID string, limit int) ([]RobloxGamePlaytime, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 5
	}

	rows, err := db.QueryContext(
		ctx,
		`
SELECT
	universe_id,
	MAX(root_place_id) AS root_place_id,
	MAX(game_name) AS game_name,
	SUM(duration_seconds) AS total_seconds,
	MAX(ended_at) AS last_seen_at,
	MIN(started_at) AS created_at,
	MAX(ended_at) AS updated_at
FROM roblox_game_playtime_sessions
WHERE stream_session_id = $1
GROUP BY universe_id
ORDER BY total_seconds DESC, updated_at DESC, universe_id ASC
LIMIT $2
`,
		strings.TrimSpace(streamSessionID),
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list roblox playtime by stream session: %w", err)
	}
	defer rows.Close()

	return scanPlaytimeRows(rows)
}

func (s *RobloxPlaytimeStore) ListTopByLastCompletedStream(ctx context.Context, limit int) ([]RobloxGamePlaytime, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var streamSessionID string
	err = db.QueryRowContext(
		ctx,
		`
SELECT stream_session_id
FROM roblox_game_playtime_sessions
ORDER BY ended_at DESC, id DESC
LIMIT 1
`,
	).Scan(&streamSessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get last roblox stream session: %w", err)
	}

	return s.ListTopByStreamSession(ctx, streamSessionID, limit)
}

func scanPlaytimeRows(rows *sql.Rows) ([]RobloxGamePlaytime, error) {
	var items []RobloxGamePlaytime
	for rows.Next() {
		var item RobloxGamePlaytime
		if err := rows.Scan(
			&item.UniverseID,
			&item.RootPlaceID,
			&item.GameName,
			&item.TotalSeconds,
			&item.LastSeenAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan roblox playtime row: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate roblox playtime rows: %w", err)
	}

	return items, nil
}
