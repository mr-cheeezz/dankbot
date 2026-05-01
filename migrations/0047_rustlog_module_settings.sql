CREATE TABLE IF NOT EXISTS rustlog_module_settings (
  id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
  enabled BOOLEAN NOT NULL DEFAULT false,
  updated_by TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO rustlog_module_settings (
  id,
  enabled,
  updated_by,
  created_at,
  updated_at
)
VALUES (
  1,
  false,
  '',
  NOW(),
  NOW()
)
ON CONFLICT (id) DO NOTHING;

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
  'rustlog',
  'RustLog',
  'logging',
  'Manage RustLog channel logging directly from chat with add/remove/status commands.',
  '["!log", "!log status", "!log list", "!log add <channel>", "!log remove <channel>"]'::jsonb,
  '[
    { "id": "enabled", "label": "Enabled", "type": "boolean", "helper_text": "Enable RustLog chat commands for moderators and broadcaster." }
  ]'::jsonb,
  80
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

