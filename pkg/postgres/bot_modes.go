package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type BotMode struct {
	ModeKey                string
	Title                  string
	Description            string
	KeywordName            string
	KeywordDescription     string
	KeywordResponse        string
	CoordinatedTwitchTitle string
	IsBuiltin              bool
	TimerEnabled           bool
	TimerMessage           string
	TimerIntervalSeconds   int
	LastTimerSentAt        time.Time
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type BotModeStore struct {
	client *Client
}

func NewBotModeStore(client *Client) *BotModeStore {
	return &BotModeStore{client: client}
}

func (s *BotModeStore) Save(ctx context.Context, mode BotMode) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	mode.ModeKey = strings.TrimSpace(strings.ToLower(mode.ModeKey))
	mode.Title = strings.TrimSpace(mode.Title)
	mode.Description = strings.TrimSpace(mode.Description)
	mode.KeywordName = strings.TrimSpace(mode.KeywordName)
	mode.KeywordDescription = strings.TrimSpace(mode.KeywordDescription)
	mode.KeywordResponse = strings.TrimSpace(mode.KeywordResponse)
	mode.CoordinatedTwitchTitle = strings.TrimSpace(mode.CoordinatedTwitchTitle)

	if mode.ModeKey == "" {
		return fmt.Errorf("mode key is required")
	}
	if mode.Title == "" {
		return fmt.Errorf("mode title is required")
	}

	_, err = db.ExecContext(
		ctx,
		`
INSERT INTO bot_modes (
	mode_key,
	title,
	description,
	keyword_name,
	keyword_description,
	keyword_response,
	coordinated_twitch_title,
	is_builtin,
	timer_enabled,
	timer_message,
	timer_interval_seconds,
	created_at,
	updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
ON CONFLICT (mode_key) DO UPDATE SET
	title = EXCLUDED.title,
	description = EXCLUDED.description,
	keyword_name = EXCLUDED.keyword_name,
	keyword_description = EXCLUDED.keyword_description,
	keyword_response = EXCLUDED.keyword_response,
	coordinated_twitch_title = EXCLUDED.coordinated_twitch_title,
	is_builtin = EXCLUDED.is_builtin,
	timer_enabled = EXCLUDED.timer_enabled,
	timer_message = EXCLUDED.timer_message,
	timer_interval_seconds = EXCLUDED.timer_interval_seconds,
	updated_at = NOW()
`,
		mode.ModeKey,
		mode.Title,
		mode.Description,
		mode.KeywordName,
		mode.KeywordDescription,
		mode.KeywordResponse,
		mode.CoordinatedTwitchTitle,
		mode.IsBuiltin,
		mode.TimerEnabled,
		mode.TimerMessage,
		mode.TimerIntervalSeconds,
	)
	if err != nil {
		return fmt.Errorf("save bot mode %q: %w", mode.ModeKey, err)
	}

	return nil
}

func (s *BotModeStore) Get(ctx context.Context, modeKey string) (*BotMode, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	modeKey = strings.TrimSpace(strings.ToLower(modeKey))
	if modeKey == "" {
		return nil, nil
	}

	var mode BotMode
	var lastTimerSentAt sql.NullTime
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	mode_key,
	title,
	description,
	keyword_name,
	keyword_description,
	keyword_response,
	coordinated_twitch_title,
	is_builtin,
	timer_enabled,
	timer_message,
	timer_interval_seconds,
	last_timer_sent_at,
	created_at,
	updated_at
FROM bot_modes
WHERE mode_key = $1
`,
		modeKey,
	).Scan(
		&mode.ModeKey,
		&mode.Title,
		&mode.Description,
		&mode.KeywordName,
		&mode.KeywordDescription,
		&mode.KeywordResponse,
		&mode.CoordinatedTwitchTitle,
		&mode.IsBuiltin,
		&mode.TimerEnabled,
		&mode.TimerMessage,
		&mode.TimerIntervalSeconds,
		&lastTimerSentAt,
		&mode.CreatedAt,
		&mode.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get bot mode %q: %w", modeKey, err)
	}

	if lastTimerSentAt.Valid {
		mode.LastTimerSentAt = lastTimerSentAt.Time
	}

	return &mode, nil
}

func (s *BotModeStore) List(ctx context.Context) ([]BotMode, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(
		ctx,
		`
SELECT
	mode_key,
	title,
	description,
	keyword_name,
	keyword_description,
	keyword_response,
	coordinated_twitch_title,
	is_builtin,
	timer_enabled,
	timer_message,
	timer_interval_seconds,
	last_timer_sent_at,
	created_at,
	updated_at
FROM bot_modes
ORDER BY is_builtin DESC, title ASC, mode_key ASC
`,
	)
	if err != nil {
		return nil, fmt.Errorf("list bot modes: %w", err)
	}
	defer rows.Close()

	var modes []BotMode
	for rows.Next() {
		var mode BotMode
		var lastTimerSentAt sql.NullTime
		if err := rows.Scan(
			&mode.ModeKey,
			&mode.Title,
			&mode.Description,
			&mode.KeywordName,
			&mode.KeywordDescription,
			&mode.KeywordResponse,
			&mode.CoordinatedTwitchTitle,
			&mode.IsBuiltin,
			&mode.TimerEnabled,
			&mode.TimerMessage,
			&mode.TimerIntervalSeconds,
			&lastTimerSentAt,
			&mode.CreatedAt,
			&mode.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan bot mode: %w", err)
		}

		if lastTimerSentAt.Valid {
			mode.LastTimerSentAt = lastTimerSentAt.Time
		}

		modes = append(modes, mode)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bot modes: %w", err)
	}

	return modes, nil
}

func (s *BotModeStore) MarkTimerSent(ctx context.Context, modeKey string, sentAt time.Time) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	modeKey = strings.TrimSpace(strings.ToLower(modeKey))
	if modeKey == "" {
		return fmt.Errorf("mode key is required")
	}

	if _, err := db.ExecContext(
		ctx,
		`UPDATE bot_modes SET last_timer_sent_at = $2, updated_at = NOW() WHERE mode_key = $1`,
		modeKey,
		sentAt,
	); err != nil {
		return fmt.Errorf("mark bot mode timer sent %q: %w", modeKey, err)
	}

	return nil
}

func (s *BotModeStore) EnsureDefaults(ctx context.Context, defaults []BotMode) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	for _, mode := range defaults {
		mode.ModeKey = strings.TrimSpace(strings.ToLower(mode.ModeKey))
		mode.Title = strings.TrimSpace(mode.Title)
		mode.Description = strings.TrimSpace(mode.Description)
		mode.KeywordName = strings.TrimSpace(mode.KeywordName)
		mode.KeywordDescription = strings.TrimSpace(mode.KeywordDescription)
		mode.KeywordResponse = strings.TrimSpace(mode.KeywordResponse)
		mode.CoordinatedTwitchTitle = strings.TrimSpace(mode.CoordinatedTwitchTitle)
		if mode.ModeKey == "" || mode.Title == "" {
			continue
		}

		if _, err := db.ExecContext(
			ctx,
			`
INSERT INTO bot_modes (
	mode_key,
	title,
	description,
	keyword_name,
	keyword_description,
	keyword_response,
	coordinated_twitch_title,
	is_builtin,
	timer_enabled,
	timer_message,
	timer_interval_seconds,
	created_at,
	updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
ON CONFLICT (mode_key) DO NOTHING
`,
			mode.ModeKey,
			mode.Title,
			mode.Description,
			mode.KeywordName,
			mode.KeywordDescription,
			mode.KeywordResponse,
			mode.CoordinatedTwitchTitle,
			mode.IsBuiltin,
			mode.TimerEnabled,
			mode.TimerMessage,
			mode.TimerIntervalSeconds,
		); err != nil {
			return fmt.Errorf("ensure default bot mode %q: %w", mode.ModeKey, err)
		}

		current, err := s.Get(ctx, mode.ModeKey)
		if err != nil {
			return err
		}
		if current == nil {
			continue
		}

		next := *current
		changed := false
		if !next.IsBuiltin && mode.IsBuiltin {
			next.IsBuiltin = true
			changed = true
		}
		if strings.TrimSpace(next.Title) == "" {
			next.Title = mode.Title
			changed = true
		}
		if strings.TrimSpace(next.Description) == "" {
			next.Description = mode.Description
			changed = true
		}
		if strings.TrimSpace(next.KeywordName) == "" {
			next.KeywordName = mode.KeywordName
			changed = true
		}
		if strings.TrimSpace(next.KeywordDescription) == "" {
			next.KeywordDescription = mode.KeywordDescription
			changed = true
		}
		if strings.TrimSpace(next.KeywordResponse) == "" {
			next.KeywordResponse = mode.KeywordResponse
			changed = true
		}
		if strings.TrimSpace(next.TimerMessage) == "" {
			next.TimerMessage = mode.TimerMessage
			changed = true
		}
		if next.TimerIntervalSeconds <= 0 {
			next.TimerIntervalSeconds = mode.TimerIntervalSeconds
			changed = true
		}

		if changed {
			if err := s.Save(ctx, next); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *BotModeStore) Delete(ctx context.Context, modeKey string) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	modeKey = strings.TrimSpace(strings.ToLower(modeKey))
	if modeKey == "" {
		return fmt.Errorf("mode key is required")
	}

	if _, err := db.ExecContext(
		ctx,
		`DELETE FROM bot_modes WHERE mode_key = $1`,
		modeKey,
	); err != nil {
		return fmt.Errorf("delete bot mode %q: %w", modeKey, err)
	}

	return nil
}
