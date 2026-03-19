CREATE TABLE IF NOT EXISTS module_catalog (
  id TEXT PRIMARY KEY,
  display_name TEXT NOT NULL,
  state TEXT NOT NULL DEFAULT '',
  detail TEXT NOT NULL DEFAULT '',
  commands JSONB NOT NULL DEFAULT '[]'::jsonb,
  settings_schema JSONB NOT NULL DEFAULT '[]'::jsonb,
  sort_order INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO module_catalog (
  id,
  display_name,
  state,
  detail,
  commands,
  settings_schema,
  sort_order
)
VALUES
  (
    'auto-followers-only',
    'Auto Followers-Only',
    'automation',
    'Watches Twitch followers-only mode and turns it back off after a set amount of time.',
    '[]'::jsonb,
    '[
      {
        "id": "auto-disable-minutes",
        "label": "Auto-disable after",
        "type": "number",
        "helper_text": "How many minutes followers-only can stay enabled before DankBot turns it back off."
      }
    ]'::jsonb,
    10
  ),
  (
    'new-chatter-greeting',
    'New Chatter Greeting',
    'engagement',
    'Greets first-time chatters with custom welcome messages you control.',
    '[]'::jsonb,
    '[
      {
        "id": "greeting-messages",
        "label": "Greeting messages (one per line)",
        "type": "textarea",
        "helper_text": "DankBot picks one line for a first-time chatter. Supports {user}, {display_name}, {login}, and {channel}."
      }
    ]'::jsonb,
    20
  ),
  (
    'game',
    'Game',
    'live',
    'Twitch game tracking plus Roblox experience resolution and playtime.',
    '["!game", "!setgame", "!gamesplayed", "!playtime"]'::jsonb,
    '[
      {
        "id": "viewer-question-enabled",
        "label": "Viewer question keyword enabled",
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
      }
    ]'::jsonb,
    30
  ),
  (
    'now-playing',
    'Now Playing',
    'live',
    'Spotify queue control, playback display, and song question handling.',
    '["!song", "!song next", "!song last", "!song add", "!song skip"]'::jsonb,
    '[
      {
        "id": "viewer-question-ai-detection",
        "label": "Use AI intent detection",
        "type": "boolean",
        "helper_text": "Helps avoid false positives when chat mentions songs without asking."
      },
      {
        "id": "viewer-question-response",
        "label": "Viewer question response",
        "type": "textarea",
        "helper_text": "Supports @{target}, {target}, and {streamer} placeholders."
      }
    ]'::jsonb,
    40
  ),
  (
    'quotes',
    'Quotes',
    'library',
    'Website-managed quote storage with lookup, add, edit, and delete.',
    '["!quote", "!add quote", "!edit quote", "!del quote"]'::jsonb,
    '[]'::jsonb,
    50
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
