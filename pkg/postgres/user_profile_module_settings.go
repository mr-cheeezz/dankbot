package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type UserProfileModuleSettings struct {
	Enabled              bool
	ShowTabSection       bool
	ShowTabHistory       bool
	ShowRedemption       bool
	ShowPollStats        bool
	ShowPredictionStats  bool
	ShowLastSeen         bool
	ShowLastChatActivity bool
	UpdatedBy            string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type UserProfileModuleSettingsStore struct {
	client *Client
}

func NewUserProfileModuleSettingsStore(client *Client) *UserProfileModuleSettingsStore {
	return &UserProfileModuleSettingsStore{client: client}
}

func DefaultUserProfileModuleSettings() UserProfileModuleSettings {
	return UserProfileModuleSettings{
		Enabled:              true,
		ShowTabSection:       true,
		ShowTabHistory:       true,
		ShowRedemption:       true,
		ShowPollStats:        true,
		ShowPredictionStats:  true,
		ShowLastSeen:         true,
		ShowLastChatActivity: true,
	}
}

func (s *UserProfileModuleSettingsStore) EnsureDefault(ctx context.Context) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	defaults := DefaultUserProfileModuleSettings()
	_, err = db.ExecContext(
		ctx,
		`
INSERT INTO user_profile_module_settings (
	id,
	enabled,
	show_tab_section,
	show_tab_history,
	show_redemption_activity,
	show_poll_stats,
	show_prediction_stats,
	show_last_seen,
	show_last_chat_activity,
	updated_by,
	created_at,
	updated_at
)
VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, '', NOW(), NOW())
ON CONFLICT (id) DO NOTHING
`,
		defaults.Enabled,
		defaults.ShowTabSection,
		defaults.ShowTabHistory,
		defaults.ShowRedemption,
		defaults.ShowPollStats,
		defaults.ShowPredictionStats,
		defaults.ShowLastSeen,
		defaults.ShowLastChatActivity,
	)
	if err != nil {
		return fmt.Errorf("ensure user profile module settings defaults: %w", err)
	}
	return nil
}

func (s *UserProfileModuleSettingsStore) Get(ctx context.Context) (*UserProfileModuleSettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var settings UserProfileModuleSettings
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	enabled,
	show_tab_section,
	show_tab_history,
	show_redemption_activity,
	show_poll_stats,
	show_prediction_stats,
	show_last_seen,
	show_last_chat_activity,
	updated_by,
	created_at,
	updated_at
FROM user_profile_module_settings
WHERE id = 1
`,
	).Scan(
		&settings.Enabled,
		&settings.ShowTabSection,
		&settings.ShowTabHistory,
		&settings.ShowRedemption,
		&settings.ShowPollStats,
		&settings.ShowPredictionStats,
		&settings.ShowLastSeen,
		&settings.ShowLastChatActivity,
		&settings.UpdatedBy,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user profile module settings: %w", err)
	}
	return &settings, nil
}

func (s *UserProfileModuleSettingsStore) Update(ctx context.Context, settings UserProfileModuleSettings) (*UserProfileModuleSettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var updated UserProfileModuleSettings
	err = db.QueryRowContext(
		ctx,
		`
UPDATE user_profile_module_settings
SET
	enabled = $1,
	show_tab_section = $2,
	show_tab_history = $3,
	show_redemption_activity = $4,
	show_poll_stats = $5,
	show_prediction_stats = $6,
	show_last_seen = $7,
	show_last_chat_activity = $8,
	updated_by = $9,
	updated_at = NOW()
WHERE id = 1
RETURNING
	enabled,
	show_tab_section,
	show_tab_history,
	show_redemption_activity,
	show_poll_stats,
	show_prediction_stats,
	show_last_seen,
	show_last_chat_activity,
	updated_by,
	created_at,
	updated_at
`,
		settings.Enabled,
		settings.ShowTabSection,
		settings.ShowTabHistory,
		settings.ShowRedemption,
		settings.ShowPollStats,
		settings.ShowPredictionStats,
		settings.ShowLastSeen,
		settings.ShowLastChatActivity,
		strings.TrimSpace(settings.UpdatedBy),
	).Scan(
		&updated.Enabled,
		&updated.ShowTabSection,
		&updated.ShowTabHistory,
		&updated.ShowRedemption,
		&updated.ShowPollStats,
		&updated.ShowPredictionStats,
		&updated.ShowLastSeen,
		&updated.ShowLastChatActivity,
		&updated.UpdatedBy,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("update user profile module settings: %w", err)
	}
	return &updated, nil
}
