package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type StreamGamePlaytime struct {
	GameKey         string
	Source          string
	TwitchGameID    string
	RobloxUniverseID int64
	GameName        string
	TotalSeconds    int64
	LastSeenAt      time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type StreamGamePlaytimeStore struct {
	client *Client
}

func NewStreamGamePlaytimeStore(client *Client) *StreamGamePlaytimeStore {
	return &StreamGamePlaytimeStore{client: client}
}

func (s *StreamGamePlaytimeStore) AddDuration(ctx context.Context, streamSessionID, gameKey, source, twitchGameID string, robloxUniverseID int64, gameName string, startedAt, endedAt time.Time) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	gameKey = strings.TrimSpace(gameKey)
	source = strings.TrimSpace(source)
	duration := endedAt.Sub(startedAt)
	seconds := int64(duration / time.Second)
	if gameKey == "" || source == "" || seconds <= 0 {
		return nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin stream game playtime transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(
		ctx,
		`
INSERT INTO stream_game_playtime (
	game_key,
	source,
	twitch_game_id,
	roblox_universe_id,
	game_name,
	total_seconds,
	last_seen_at,
	created_at,
	updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
ON CONFLICT (game_key) DO UPDATE SET
	source = EXCLUDED.source,
	twitch_game_id = EXCLUDED.twitch_game_id,
	roblox_universe_id = EXCLUDED.roblox_universe_id,
	game_name = CASE
		WHEN EXCLUDED.game_name <> '' THEN EXCLUDED.game_name
		ELSE stream_game_playtime.game_name
	END,
	total_seconds = stream_game_playtime.total_seconds + EXCLUDED.total_seconds,
	last_seen_at = EXCLUDED.last_seen_at,
	updated_at = NOW()
`,
		gameKey,
		source,
		strings.TrimSpace(twitchGameID),
		robloxUniverseID,
		strings.TrimSpace(gameName),
		seconds,
		endedAt.UTC(),
	); err != nil {
		return fmt.Errorf("add stream game playtime %q: %w", gameKey, err)
	}

	if strings.TrimSpace(streamSessionID) != "" {
		if _, err = tx.ExecContext(
			ctx,
			`
INSERT INTO stream_game_playtime_sessions (
	stream_session_id,
	game_key,
	source,
	twitch_game_id,
	roblox_universe_id,
	game_name,
	started_at,
	ended_at,
	duration_seconds,
	created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
`,
			strings.TrimSpace(streamSessionID),
			gameKey,
			source,
			strings.TrimSpace(twitchGameID),
			robloxUniverseID,
			strings.TrimSpace(gameName),
			startedAt.UTC(),
			endedAt.UTC(),
			seconds,
		); err != nil {
			return fmt.Errorf("insert stream game playtime session %q: %w", gameKey, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit stream game playtime transaction: %w", err)
	}

	return nil
}

func (s *StreamGamePlaytimeStore) Get(ctx context.Context, gameKey string) (*StreamGamePlaytime, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var item StreamGamePlaytime
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	game_key,
	source,
	twitch_game_id,
	roblox_universe_id,
	game_name,
	total_seconds,
	last_seen_at,
	created_at,
	updated_at
FROM stream_game_playtime
WHERE game_key = $1
`,
		strings.TrimSpace(gameKey),
	).Scan(
		&item.GameKey,
		&item.Source,
		&item.TwitchGameID,
		&item.RobloxUniverseID,
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
		return nil, fmt.Errorf("get stream game playtime %q: %w", gameKey, err)
	}

	return &item, nil
}

func (s *StreamGamePlaytimeStore) ListTop(ctx context.Context, limit int) ([]StreamGamePlaytime, error) {
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
	game_key,
	source,
	twitch_game_id,
	roblox_universe_id,
	game_name,
	total_seconds,
	last_seen_at,
	created_at,
	updated_at
FROM stream_game_playtime
ORDER BY total_seconds DESC, updated_at DESC, game_key ASC
LIMIT $1
`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list top stream game playtime: %w", err)
	}
	defer rows.Close()

	return scanStreamGamePlaytimeRows(rows)
}

func (s *StreamGamePlaytimeStore) ListTopByRange(ctx context.Context, start, end time.Time, limit int) ([]StreamGamePlaytime, error) {
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
	game_key,
	MAX(source) AS source,
	MAX(twitch_game_id) AS twitch_game_id,
	MAX(roblox_universe_id) AS roblox_universe_id,
	MAX(game_name) AS game_name,
	SUM(duration_seconds) AS total_seconds,
	MAX(ended_at) AS last_seen_at,
	MIN(started_at) AS created_at,
	MAX(ended_at) AS updated_at
FROM stream_game_playtime_sessions
WHERE ended_at >= $1 AND ended_at < $2
GROUP BY game_key
ORDER BY total_seconds DESC, updated_at DESC, game_key ASC
LIMIT $3
`,
		start.UTC(),
		end.UTC(),
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list stream game playtime by range: %w", err)
	}
	defer rows.Close()

	return scanStreamGamePlaytimeRows(rows)
}

func (s *StreamGamePlaytimeStore) ListTopByStreamSession(ctx context.Context, streamSessionID string, limit int) ([]StreamGamePlaytime, error) {
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
	game_key,
	MAX(source) AS source,
	MAX(twitch_game_id) AS twitch_game_id,
	MAX(roblox_universe_id) AS roblox_universe_id,
	MAX(game_name) AS game_name,
	SUM(duration_seconds) AS total_seconds,
	MAX(ended_at) AS last_seen_at,
	MIN(started_at) AS created_at,
	MAX(ended_at) AS updated_at
FROM stream_game_playtime_sessions
WHERE stream_session_id = $1
GROUP BY game_key
ORDER BY total_seconds DESC, updated_at DESC, game_key ASC
LIMIT $2
`,
		strings.TrimSpace(streamSessionID),
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list stream game playtime by stream session: %w", err)
	}
	defer rows.Close()

	return scanStreamGamePlaytimeRows(rows)
}

func (s *StreamGamePlaytimeStore) ListTopByLastCompletedStream(ctx context.Context, limit int) ([]StreamGamePlaytime, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var streamSessionID string
	err = db.QueryRowContext(
		ctx,
		`
SELECT stream_session_id
FROM stream_game_playtime_sessions
ORDER BY ended_at DESC, id DESC
LIMIT 1
`,
	).Scan(&streamSessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get last completed stream session: %w", err)
	}

	return s.ListTopByStreamSession(ctx, streamSessionID, limit)
}

func (s *StreamGamePlaytimeStore) LastCompletedStreamEndedAt(ctx context.Context) (*time.Time, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var endedAt time.Time
	err = db.QueryRowContext(
		ctx,
		`
SELECT ended_at
FROM stream_game_playtime_sessions
ORDER BY ended_at DESC, id DESC
LIMIT 1
`,
	).Scan(&endedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get last completed stream ended at: %w", err)
	}

	return &endedAt, nil
}

func scanStreamGamePlaytimeRows(rows *sql.Rows) ([]StreamGamePlaytime, error) {
	var items []StreamGamePlaytime
	for rows.Next() {
		var item StreamGamePlaytime
		if err := rows.Scan(
			&item.GameKey,
			&item.Source,
			&item.TwitchGameID,
			&item.RobloxUniverseID,
			&item.GameName,
			&item.TotalSeconds,
			&item.LastSeenAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan stream game playtime row: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate stream game playtime rows: %w", err)
	}

	return items, nil
}
