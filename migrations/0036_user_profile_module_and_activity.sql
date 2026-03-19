CREATE TABLE IF NOT EXISTS user_profile_module_settings (
  id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
  enabled BOOLEAN NOT NULL DEFAULT true,
  show_tab_section BOOLEAN NOT NULL DEFAULT true,
  show_tab_history BOOLEAN NOT NULL DEFAULT true,
  show_redemption_activity BOOLEAN NOT NULL DEFAULT true,
  show_poll_stats BOOLEAN NOT NULL DEFAULT true,
  show_prediction_stats BOOLEAN NOT NULL DEFAULT true,
  show_last_seen BOOLEAN NOT NULL DEFAULT true,
  show_last_chat_activity BOOLEAN NOT NULL DEFAULT true,
  updated_by TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

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
VALUES (
  1,
  true,
  true,
  true,
  true,
  true,
  true,
  true,
  true,
  '',
  NOW(),
  NOW()
)
ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS user_tab_events (
  id BIGSERIAL PRIMARY KEY,
  login TEXT NOT NULL,
  action TEXT NOT NULL,
  amount_cents BIGINT NOT NULL DEFAULT 0,
  balance_cents BIGINT NOT NULL DEFAULT 0,
  note TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_tab_events_login_created_at
  ON user_tab_events (login, created_at DESC);

CREATE TABLE IF NOT EXISTS twitch_prediction_events (
  id BIGSERIAL PRIMARY KEY,
  twitch_subscription_id TEXT NOT NULL,
  event_type TEXT NOT NULL,
  prediction_id TEXT NOT NULL,
  broadcaster_user_id TEXT NOT NULL,
  broadcaster_user_login TEXT NOT NULL,
  broadcaster_user_name TEXT NOT NULL,
  title TEXT NOT NULL,
  status TEXT NOT NULL,
  winning_outcome_id TEXT NOT NULL DEFAULT '',
  started_at TIMESTAMPTZ,
  ended_at TIMESTAMPTZ,
  locked_at TIMESTAMPTZ,
  raw_event JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_twitch_prediction_events_broadcaster_created_at
  ON twitch_prediction_events (broadcaster_user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS twitch_user_chat_activity (
  user_id TEXT PRIMARY KEY,
  user_login TEXT NOT NULL,
  display_name TEXT NOT NULL DEFAULT '',
  message_count BIGINT NOT NULL DEFAULT 0,
  last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  last_chat_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_twitch_user_chat_activity_login
  ON twitch_user_chat_activity (user_login);

INSERT INTO module_catalog (
  id,
  display_name,
  state,
  detail,
  commands,
  settings_schema,
  sort_order
)
VALUES (
  'user-profile',
  'User Profile',
  'public',
  'Controls which profile stats and activity blocks are shown on /user/{twitchusername_raw}.',
  '[]'::jsonb,
  '[
    { "id": "enabled", "label": "Enabled", "type": "boolean", "helper_text": "Enable or disable public user profiles." },
    { "id": "show-tab-section", "label": "Show tab section", "type": "boolean", "helper_text": "Show tab balance and tab activity block." },
    { "id": "show-tab-history", "label": "Show tab history", "type": "boolean", "helper_text": "Allow viewing recent tab entries and full history." },
    { "id": "show-redemption-activity", "label": "Show redemption activity", "type": "boolean", "helper_text": "Show channel point summary and recent redemption items." },
    { "id": "show-poll-stats", "label": "Show poll stats", "type": "boolean", "helper_text": "Show stored poll event stats for the channel." },
    { "id": "show-prediction-stats", "label": "Show prediction stats", "type": "boolean", "helper_text": "Show stored prediction event stats for the channel." },
    { "id": "show-last-seen", "label": "Show last seen", "type": "boolean", "helper_text": "Show the latest seen timestamp." },
    { "id": "show-last-chat-activity", "label": "Show last active in chat", "type": "boolean", "helper_text": "Show the latest tracked chat activity timestamp." }
  ]'::jsonb,
  70
)
ON CONFLICT (id) DO UPDATE
SET
  display_name = EXCLUDED.display_name,
  state = EXCLUDED.state,
  detail = EXCLUDED.detail,
  commands = EXCLUDED.commands,
  settings_schema = EXCLUDED.settings_schema,
  sort_order = EXCLUDED.sort_order,
  updated_at = NOW();

UPDATE module_catalog
SET
  commands = '["!tab <user>", "!tab add <user> <amount>", "!tab set <user> <amount>", "!tab paid <user>", "!tab give <user>"]'::jsonb,
  updated_at = NOW()
WHERE id = 'tabs';
