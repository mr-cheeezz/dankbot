package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const GlobalBotStateKey = "global"

type BotState struct {
	StateKey          string
	CurrentModeKey    string
	CurrentModeParam  string
	KillswitchEnabled bool
	UpdatedBy         string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type BotStateStore struct {
	client *Client
}

func NewBotStateStore(client *Client) *BotStateStore {
	return &BotStateStore{client: client}
}

func (s *BotStateStore) Ensure(ctx context.Context, defaultModeKey string) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	defaultModeKey = strings.TrimSpace(strings.ToLower(defaultModeKey))
	if defaultModeKey == "" {
		return fmt.Errorf("default mode key is required")
	}

	_, err = db.ExecContext(
		ctx,
		`
INSERT INTO bot_state (
	state_key,
	current_mode_key,
	current_mode_param,
	killswitch_enabled,
	updated_by,
	created_at,
	updated_at
)
VALUES ($1, $2, '', FALSE, '', NOW(), NOW())
ON CONFLICT (state_key) DO NOTHING
`,
		GlobalBotStateKey,
		defaultModeKey,
	)
	if err != nil {
		return fmt.Errorf("ensure bot state: %w", err)
	}

	return nil
}

func (s *BotStateStore) Get(ctx context.Context) (*BotState, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var state BotState
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	state_key,
	current_mode_key,
	current_mode_param,
	killswitch_enabled,
	updated_by,
	created_at,
	updated_at
FROM bot_state
WHERE state_key = $1
`,
		GlobalBotStateKey,
	).Scan(
		&state.StateKey,
		&state.CurrentModeKey,
		&state.CurrentModeParam,
		&state.KillswitchEnabled,
		&state.UpdatedBy,
		&state.CreatedAt,
		&state.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get bot state: %w", err)
	}

	return &state, nil
}

func (s *BotStateStore) SetCurrentMode(ctx context.Context, modeKey, modeParam, updatedBy string) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(
		ctx,
		`
UPDATE bot_state
SET
	current_mode_key = $2,
	current_mode_param = $3,
	updated_by = $4,
	updated_at = NOW()
WHERE state_key = $1
`,
		GlobalBotStateKey,
		strings.TrimSpace(strings.ToLower(modeKey)),
		strings.TrimSpace(modeParam),
		strings.TrimSpace(updatedBy),
	)
	if err != nil {
		return fmt.Errorf("set current mode: %w", err)
	}

	return nil
}

func (s *BotStateStore) ToggleKillswitch(ctx context.Context, updatedBy string) (*BotState, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	err = s.Ensure(ctx, "join")
	if err != nil {
		return nil, err
	}

	if _, err := db.ExecContext(
		ctx,
		`
UPDATE bot_state
SET
	killswitch_enabled = NOT killswitch_enabled,
	updated_by = $2,
	updated_at = NOW()
WHERE state_key = $1
`,
		GlobalBotStateKey,
		strings.TrimSpace(updatedBy),
	); err != nil {
		return nil, fmt.Errorf("toggle killswitch: %w", err)
	}

	return s.Get(ctx)
}
