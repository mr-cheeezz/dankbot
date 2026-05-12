ALTER TABLE followers_only_module_settings
  ADD COLUMN IF NOT EXISTS auto_disable_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN IF NOT EXISTS online_slow_mode_action TEXT NOT NULL DEFAULT 'no-change',
  ADD COLUMN IF NOT EXISTS online_slow_mode_seconds INTEGER NOT NULL DEFAULT 30,
  ADD COLUMN IF NOT EXISTS online_emote_mode_action TEXT NOT NULL DEFAULT 'no-change',
  ADD COLUMN IF NOT EXISTS online_unique_chat_mode_action TEXT NOT NULL DEFAULT 'no-change',
  ADD COLUMN IF NOT EXISTS online_subscriber_mode_action TEXT NOT NULL DEFAULT 'no-change',
  ADD COLUMN IF NOT EXISTS online_follower_mode_action TEXT NOT NULL DEFAULT 'no-change',
  ADD COLUMN IF NOT EXISTS online_follower_mode_minutes INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS offline_slow_mode_action TEXT NOT NULL DEFAULT 'no-change',
  ADD COLUMN IF NOT EXISTS offline_slow_mode_seconds INTEGER NOT NULL DEFAULT 30,
  ADD COLUMN IF NOT EXISTS offline_emote_mode_action TEXT NOT NULL DEFAULT 'no-change',
  ADD COLUMN IF NOT EXISTS offline_unique_chat_mode_action TEXT NOT NULL DEFAULT 'no-change',
  ADD COLUMN IF NOT EXISTS offline_subscriber_mode_action TEXT NOT NULL DEFAULT 'no-change',
  ADD COLUMN IF NOT EXISTS offline_follower_mode_action TEXT NOT NULL DEFAULT 'no-change',
  ADD COLUMN IF NOT EXISTS offline_follower_mode_minutes INTEGER NOT NULL DEFAULT 0;

UPDATE followers_only_module_settings
SET
  auto_disable_enabled = COALESCE(auto_disable_enabled, TRUE),
  online_slow_mode_action = COALESCE(NULLIF(TRIM(online_slow_mode_action), ''), 'no-change'),
  online_slow_mode_seconds = CASE WHEN online_slow_mode_seconds IS NULL OR online_slow_mode_seconds < 0 THEN 30 ELSE online_slow_mode_seconds END,
  online_emote_mode_action = COALESCE(NULLIF(TRIM(online_emote_mode_action), ''), 'no-change'),
  online_unique_chat_mode_action = COALESCE(NULLIF(TRIM(online_unique_chat_mode_action), ''), 'no-change'),
  online_subscriber_mode_action = COALESCE(NULLIF(TRIM(online_subscriber_mode_action), ''), 'no-change'),
  online_follower_mode_action = COALESCE(NULLIF(TRIM(online_follower_mode_action), ''), 'no-change'),
  online_follower_mode_minutes = CASE WHEN online_follower_mode_minutes IS NULL OR online_follower_mode_minutes < 0 THEN 0 ELSE online_follower_mode_minutes END,
  offline_slow_mode_action = COALESCE(NULLIF(TRIM(offline_slow_mode_action), ''), 'no-change'),
  offline_slow_mode_seconds = CASE WHEN offline_slow_mode_seconds IS NULL OR offline_slow_mode_seconds < 0 THEN 30 ELSE offline_slow_mode_seconds END,
  offline_emote_mode_action = COALESCE(NULLIF(TRIM(offline_emote_mode_action), ''), 'no-change'),
  offline_unique_chat_mode_action = COALESCE(NULLIF(TRIM(offline_unique_chat_mode_action), ''), 'no-change'),
  offline_subscriber_mode_action = COALESCE(NULLIF(TRIM(offline_subscriber_mode_action), ''), 'no-change'),
  offline_follower_mode_action = COALESCE(NULLIF(TRIM(offline_follower_mode_action), ''), 'no-change'),
  offline_follower_mode_minutes = CASE WHEN offline_follower_mode_minutes IS NULL OR offline_follower_mode_minutes < 0 THEN 0 ELSE offline_follower_mode_minutes END
WHERE id = 1;

ALTER TABLE game_module_settings
  ADD COLUMN IF NOT EXISTS enabled BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN IF NOT EXISTS game_command_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN IF NOT EXISTS playtime_command_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN IF NOT EXISTS gamesplayed_command_enabled BOOLEAN NOT NULL DEFAULT TRUE;

UPDATE game_module_settings
SET
  enabled = COALESCE(enabled, TRUE),
  game_command_enabled = COALESCE(game_command_enabled, TRUE),
  playtime_command_enabled = COALESCE(playtime_command_enabled, TRUE),
  gamesplayed_command_enabled = COALESCE(gamesplayed_command_enabled, TRUE)
WHERE id = 1;

UPDATE module_catalog
SET
  display_name = 'Auto Chat States',
  detail = 'Applies Twitch chat state presets when the stream goes online or offline, and can still auto-turn follower-only back off later.',
  settings_schema = '[
    {
      "id": "online-slow-mode-action",
      "label": "Online: Slow mode",
      "type": "select",
      "helper_text": "What should happen to slow mode when the stream comes online?",
      "options": ["no-change", "disable", "enable"]
    },
    {
      "id": "online-slow-mode-seconds",
      "label": "Online: Slow mode seconds",
      "type": "number",
      "helper_text": "Used when online slow mode is enabled."
    },
    {
      "id": "online-emote-mode-action",
      "label": "Online: Emote-only",
      "type": "select",
      "helper_text": "What should happen to emote-only mode when the stream comes online?",
      "options": ["no-change", "disable", "enable"]
    },
    {
      "id": "online-unique-chat-mode-action",
      "label": "Online: Unique chat (R9K)",
      "type": "select",
      "helper_text": "What should happen to unique chat mode when the stream comes online?",
      "options": ["no-change", "disable", "enable"]
    },
    {
      "id": "online-subscriber-mode-action",
      "label": "Online: Subscribers-only",
      "type": "select",
      "helper_text": "What should happen to subscribers-only mode when the stream comes online?",
      "options": ["no-change", "disable", "enable"]
    },
    {
      "id": "online-follower-mode-action",
      "label": "Online: Followers-only",
      "type": "select",
      "helper_text": "What should happen to followers-only mode when the stream comes online?",
      "options": ["no-change", "disable", "enable"]
    },
    {
      "id": "online-follower-mode-minutes",
      "label": "Online: Followers-only minutes",
      "type": "number",
      "helper_text": "Used when online followers-only is enabled. Use 0 for any follower age."
    },
    {
      "id": "offline-slow-mode-action",
      "label": "Offline: Slow mode",
      "type": "select",
      "helper_text": "What should happen to slow mode when the stream goes offline?",
      "options": ["no-change", "disable", "enable"]
    },
    {
      "id": "offline-slow-mode-seconds",
      "label": "Offline: Slow mode seconds",
      "type": "number",
      "helper_text": "Used when offline slow mode is enabled."
    },
    {
      "id": "offline-emote-mode-action",
      "label": "Offline: Emote-only",
      "type": "select",
      "helper_text": "What should happen to emote-only mode when the stream goes offline?",
      "options": ["no-change", "disable", "enable"]
    },
    {
      "id": "offline-unique-chat-mode-action",
      "label": "Offline: Unique chat (R9K)",
      "type": "select",
      "helper_text": "What should happen to unique chat mode when the stream goes offline?",
      "options": ["no-change", "disable", "enable"]
    },
    {
      "id": "offline-subscriber-mode-action",
      "label": "Offline: Subscribers-only",
      "type": "select",
      "helper_text": "What should happen to subscribers-only mode when the stream goes offline?",
      "options": ["no-change", "disable", "enable"]
    },
    {
      "id": "offline-follower-mode-action",
      "label": "Offline: Followers-only",
      "type": "select",
      "helper_text": "What should happen to followers-only mode when the stream goes offline?",
      "options": ["no-change", "disable", "enable"]
    },
    {
      "id": "offline-follower-mode-minutes",
      "label": "Offline: Followers-only minutes",
      "type": "number",
      "helper_text": "Used when offline followers-only is enabled. Use 0 for any follower age."
    },
    {
      "id": "auto-disable-enabled",
      "label": "Auto-turn follower-only back off",
      "type": "boolean",
      "helper_text": "After follower-only has been active for a while, DankBot can turn it back off automatically."
    },
    {
      "id": "auto-disable-minutes",
      "label": "Follower-only auto-disable minutes",
      "type": "number",
      "helper_text": "How long follower-only can stay enabled before DankBot turns it back off."
    }
  ]'::jsonb,
  updated_at = NOW()
WHERE id = 'auto-followers-only';

UPDATE module_catalog
SET
  settings_schema = '[
    {
      "id": "viewer-question-enabled",
      "label": "Enable viewer question keyword",
      "type": "boolean",
      "helper_text": "Enable/disable the built-in what-game question keyword handling."
    },
    {
      "id": "viewer-question-ai-detection",
      "label": "Use AI intent detection",
      "type": "boolean",
      "helper_text": "Helps avoid false positives when people mention a game without asking."
    },
    {
      "id": "viewer-question-response",
      "label": "Viewer question response",
      "type": "textarea",
      "helper_text": "Supports @{target}, {target}, and {streamer} placeholders."
    },
    {
      "id": "game-command-enabled",
      "label": "Enable !game",
      "type": "boolean",
      "helper_text": "Enable/disable the !game command."
    },
    {
      "id": "playtime-command-enabled",
      "label": "Enable !playtime",
      "type": "boolean",
      "helper_text": "Enable/disable the !playtime command."
    },
    {
      "id": "gamesplayed-command-enabled",
      "label": "Enable !gamesplayed",
      "type": "boolean",
      "helper_text": "Enable/disable the !gamesplayed command."
    },
    {
      "id": "playtime-template",
      "label": "!playtime response template",
      "type": "textarea",
      "helper_text": "Supports {streamer}, {game}, and {duration}."
    },
    {
      "id": "gamesplayed-template",
      "label": "!gamesplayed response template",
      "type": "textarea",
      "helper_text": "Supports {label} and {items}."
    },
    {
      "id": "gamesplayed-item-template",
      "label": "!gamesplayed item template",
      "type": "text",
      "helper_text": "Supports {game} and {duration}."
    },
    {
      "id": "gamesplayed-limit",
      "label": "Top games limit",
      "type": "number",
      "helper_text": "How many games to include in the !gamesplayed output."
    }
  ]'::jsonb,
  updated_at = NOW()
WHERE id = 'game';
