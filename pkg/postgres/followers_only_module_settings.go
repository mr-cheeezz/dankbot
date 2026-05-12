package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type FollowersOnlyModuleSettings struct {
	Enabled                 bool
	EnabledWhenOffline      bool
	AutoDisableEnabled      bool
	AutoDisableAfterMinutes int
	OnlineSlowModeAction    string
	OnlineSlowModeSeconds   int
	OnlineEmoteModeAction   string
	OnlineUniqueChatAction  string
	OnlineSubscriberAction  string
	OnlineFollowerAction    string
	OnlineFollowerMinutes   int
	OfflineSlowModeAction   string
	OfflineSlowModeSeconds  int
	OfflineEmoteModeAction  string
	OfflineUniqueChatAction string
	OfflineSubscriberAction string
	OfflineFollowerAction   string
	OfflineFollowerMinutes  int
	UpdatedBy               string
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

type FollowersOnlyModuleSettingsStore struct {
	client *Client
}

func NewFollowersOnlyModuleSettingsStore(client *Client) *FollowersOnlyModuleSettingsStore {
	return &FollowersOnlyModuleSettingsStore{client: client}
}

func DefaultFollowersOnlyModuleSettings() FollowersOnlyModuleSettings {
	return FollowersOnlyModuleSettings{
		Enabled:                 false,
		EnabledWhenOffline:      false,
		AutoDisableEnabled:      true,
		AutoDisableAfterMinutes: 30,
		OnlineSlowModeAction:    "no-change",
		OnlineSlowModeSeconds:   30,
		OnlineEmoteModeAction:   "no-change",
		OnlineUniqueChatAction:  "no-change",
		OnlineSubscriberAction:  "no-change",
		OnlineFollowerAction:    "no-change",
		OnlineFollowerMinutes:   0,
		OfflineSlowModeAction:   "no-change",
		OfflineSlowModeSeconds:  30,
		OfflineEmoteModeAction:  "no-change",
		OfflineUniqueChatAction: "no-change",
		OfflineSubscriberAction: "no-change",
		OfflineFollowerAction:   "no-change",
		OfflineFollowerMinutes:  0,
	}
}

func (s *FollowersOnlyModuleSettingsStore) EnsureDefault(ctx context.Context) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	defaults := DefaultFollowersOnlyModuleSettings()
	_, err = db.ExecContext(
		ctx,
		`
INSERT INTO followers_only_module_settings (
	id,
	enabled,
	enabled_when_offline,
	auto_disable_enabled,
	auto_disable_after_minutes,
	online_slow_mode_action,
	online_slow_mode_seconds,
	online_emote_mode_action,
	online_unique_chat_mode_action,
	online_subscriber_mode_action,
	online_follower_mode_action,
	online_follower_mode_minutes,
	offline_slow_mode_action,
	offline_slow_mode_seconds,
	offline_emote_mode_action,
	offline_unique_chat_mode_action,
	offline_subscriber_mode_action,
	offline_follower_mode_action,
	offline_follower_mode_minutes,
	updated_by,
	created_at,
	updated_at
)
VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, '', NOW(), NOW())
ON CONFLICT (id) DO NOTHING
`,
		defaults.Enabled,
		defaults.EnabledWhenOffline,
		defaults.AutoDisableEnabled,
		defaults.AutoDisableAfterMinutes,
		defaults.OnlineSlowModeAction,
		defaults.OnlineSlowModeSeconds,
		defaults.OnlineEmoteModeAction,
		defaults.OnlineUniqueChatAction,
		defaults.OnlineSubscriberAction,
		defaults.OnlineFollowerAction,
		defaults.OnlineFollowerMinutes,
		defaults.OfflineSlowModeAction,
		defaults.OfflineSlowModeSeconds,
		defaults.OfflineEmoteModeAction,
		defaults.OfflineUniqueChatAction,
		defaults.OfflineSubscriberAction,
		defaults.OfflineFollowerAction,
		defaults.OfflineFollowerMinutes,
	)
	if err != nil {
		return fmt.Errorf("ensure followers-only module settings defaults: %w", err)
	}

	return nil
}

func (s *FollowersOnlyModuleSettingsStore) Get(ctx context.Context) (*FollowersOnlyModuleSettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var settings FollowersOnlyModuleSettings
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	enabled,
	enabled_when_offline,
	auto_disable_enabled,
	auto_disable_after_minutes,
	online_slow_mode_action,
	online_slow_mode_seconds,
	online_emote_mode_action,
	online_unique_chat_mode_action,
	online_subscriber_mode_action,
	online_follower_mode_action,
	online_follower_mode_minutes,
	offline_slow_mode_action,
	offline_slow_mode_seconds,
	offline_emote_mode_action,
	offline_unique_chat_mode_action,
	offline_subscriber_mode_action,
	offline_follower_mode_action,
	offline_follower_mode_minutes,
	updated_by,
	created_at,
	updated_at
FROM followers_only_module_settings
WHERE id = 1
`,
	).Scan(
		&settings.Enabled,
		&settings.EnabledWhenOffline,
		&settings.AutoDisableEnabled,
		&settings.AutoDisableAfterMinutes,
		&settings.OnlineSlowModeAction,
		&settings.OnlineSlowModeSeconds,
		&settings.OnlineEmoteModeAction,
		&settings.OnlineUniqueChatAction,
		&settings.OnlineSubscriberAction,
		&settings.OnlineFollowerAction,
		&settings.OnlineFollowerMinutes,
		&settings.OfflineSlowModeAction,
		&settings.OfflineSlowModeSeconds,
		&settings.OfflineEmoteModeAction,
		&settings.OfflineUniqueChatAction,
		&settings.OfflineSubscriberAction,
		&settings.OfflineFollowerAction,
		&settings.OfflineFollowerMinutes,
		&settings.UpdatedBy,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get followers-only module settings: %w", err)
	}

	settings = normalizeFollowersOnlyModuleSettings(settings)
	return &settings, nil
}

func (s *FollowersOnlyModuleSettingsStore) Update(ctx context.Context, settings FollowersOnlyModuleSettings) (*FollowersOnlyModuleSettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var updated FollowersOnlyModuleSettings
	err = db.QueryRowContext(
		ctx,
		`
UPDATE followers_only_module_settings
SET
	enabled = $1,
	enabled_when_offline = $2,
	auto_disable_enabled = $3,
	auto_disable_after_minutes = $4,
	online_slow_mode_action = $5,
	online_slow_mode_seconds = $6,
	online_emote_mode_action = $7,
	online_unique_chat_mode_action = $8,
	online_subscriber_mode_action = $9,
	online_follower_mode_action = $10,
	online_follower_mode_minutes = $11,
	offline_slow_mode_action = $12,
	offline_slow_mode_seconds = $13,
	offline_emote_mode_action = $14,
	offline_unique_chat_mode_action = $15,
	offline_subscriber_mode_action = $16,
	offline_follower_mode_action = $17,
	offline_follower_mode_minutes = $18,
	updated_by = $19,
	updated_at = NOW()
WHERE id = 1
RETURNING
	enabled,
	enabled_when_offline,
	auto_disable_enabled,
	auto_disable_after_minutes,
	online_slow_mode_action,
	online_slow_mode_seconds,
	online_emote_mode_action,
	online_unique_chat_mode_action,
	online_subscriber_mode_action,
	online_follower_mode_action,
	online_follower_mode_minutes,
	offline_slow_mode_action,
	offline_slow_mode_seconds,
	offline_emote_mode_action,
	offline_unique_chat_mode_action,
	offline_subscriber_mode_action,
	offline_follower_mode_action,
	offline_follower_mode_minutes,
	updated_by,
	created_at,
	updated_at
`,
		settings.Enabled,
		settings.EnabledWhenOffline,
		settings.AutoDisableEnabled,
		normalizeFollowersOnlyAutoDisableMinutes(settings.AutoDisableAfterMinutes),
		normalizeFollowersOnlyModeAction(settings.OnlineSlowModeAction),
		normalizeFollowersOnlySlowModeSeconds(settings.OnlineSlowModeSeconds),
		normalizeFollowersOnlyModeAction(settings.OnlineEmoteModeAction),
		normalizeFollowersOnlyModeAction(settings.OnlineUniqueChatAction),
		normalizeFollowersOnlyModeAction(settings.OnlineSubscriberAction),
		normalizeFollowersOnlyModeAction(settings.OnlineFollowerAction),
		normalizeFollowersOnlyFollowerMinutes(settings.OnlineFollowerMinutes),
		normalizeFollowersOnlyModeAction(settings.OfflineSlowModeAction),
		normalizeFollowersOnlySlowModeSeconds(settings.OfflineSlowModeSeconds),
		normalizeFollowersOnlyModeAction(settings.OfflineEmoteModeAction),
		normalizeFollowersOnlyModeAction(settings.OfflineUniqueChatAction),
		normalizeFollowersOnlyModeAction(settings.OfflineSubscriberAction),
		normalizeFollowersOnlyModeAction(settings.OfflineFollowerAction),
		normalizeFollowersOnlyFollowerMinutes(settings.OfflineFollowerMinutes),
		strings.TrimSpace(settings.UpdatedBy),
	).Scan(
		&updated.Enabled,
		&updated.EnabledWhenOffline,
		&updated.AutoDisableEnabled,
		&updated.AutoDisableAfterMinutes,
		&updated.OnlineSlowModeAction,
		&updated.OnlineSlowModeSeconds,
		&updated.OnlineEmoteModeAction,
		&updated.OnlineUniqueChatAction,
		&updated.OnlineSubscriberAction,
		&updated.OnlineFollowerAction,
		&updated.OnlineFollowerMinutes,
		&updated.OfflineSlowModeAction,
		&updated.OfflineSlowModeSeconds,
		&updated.OfflineEmoteModeAction,
		&updated.OfflineUniqueChatAction,
		&updated.OfflineSubscriberAction,
		&updated.OfflineFollowerAction,
		&updated.OfflineFollowerMinutes,
		&updated.UpdatedBy,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("update followers-only module settings: %w", err)
	}

	updated = normalizeFollowersOnlyModuleSettings(updated)
	return &updated, nil
}

func normalizeFollowersOnlyModuleSettings(settings FollowersOnlyModuleSettings) FollowersOnlyModuleSettings {
	settings.AutoDisableAfterMinutes = normalizeFollowersOnlyAutoDisableMinutes(settings.AutoDisableAfterMinutes)
	settings.OnlineSlowModeAction = normalizeFollowersOnlyModeAction(settings.OnlineSlowModeAction)
	settings.OnlineSlowModeSeconds = normalizeFollowersOnlySlowModeSeconds(settings.OnlineSlowModeSeconds)
	settings.OnlineEmoteModeAction = normalizeFollowersOnlyModeAction(settings.OnlineEmoteModeAction)
	settings.OnlineUniqueChatAction = normalizeFollowersOnlyModeAction(settings.OnlineUniqueChatAction)
	settings.OnlineSubscriberAction = normalizeFollowersOnlyModeAction(settings.OnlineSubscriberAction)
	settings.OnlineFollowerAction = normalizeFollowersOnlyModeAction(settings.OnlineFollowerAction)
	settings.OnlineFollowerMinutes = normalizeFollowersOnlyFollowerMinutes(settings.OnlineFollowerMinutes)
	settings.OfflineSlowModeAction = normalizeFollowersOnlyModeAction(settings.OfflineSlowModeAction)
	settings.OfflineSlowModeSeconds = normalizeFollowersOnlySlowModeSeconds(settings.OfflineSlowModeSeconds)
	settings.OfflineEmoteModeAction = normalizeFollowersOnlyModeAction(settings.OfflineEmoteModeAction)
	settings.OfflineUniqueChatAction = normalizeFollowersOnlyModeAction(settings.OfflineUniqueChatAction)
	settings.OfflineSubscriberAction = normalizeFollowersOnlyModeAction(settings.OfflineSubscriberAction)
	settings.OfflineFollowerAction = normalizeFollowersOnlyModeAction(settings.OfflineFollowerAction)
	settings.OfflineFollowerMinutes = normalizeFollowersOnlyFollowerMinutes(settings.OfflineFollowerMinutes)
	return settings
}

func normalizeFollowersOnlyAutoDisableMinutes(value int) int {
	switch {
	case value < 1:
		return 30
	case value > 24*60:
		return 24 * 60
	default:
		return value
	}
}

func normalizeFollowersOnlyModeAction(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "enable":
		return "enable"
	case "disable":
		return "disable"
	default:
		return "no-change"
	}
}

func normalizeFollowersOnlySlowModeSeconds(value int) int {
	switch {
	case value <= 0:
		return 30
	case value > 120:
		return 120
	default:
		return value
	}
}

func normalizeFollowersOnlyFollowerMinutes(value int) int {
	switch {
	case value < 0:
		return 0
	case value > 129600:
		return 129600
	default:
		return value
	}
}
