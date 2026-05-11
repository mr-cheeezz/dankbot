CREATE TABLE IF NOT EXISTS custom_commands (
  id TEXT PRIMARY KEY,
  platform TEXT NOT NULL,
  name TEXT NOT NULL,
  aliases_json JSONB NOT NULL DEFAULT '[]'::jsonb,
  group_name TEXT NOT NULL DEFAULT 'custom',
  state_label TEXT NOT NULL DEFAULT 'enabled',
  description TEXT NOT NULL DEFAULT '',
  example TEXT NOT NULL DEFAULT '',
  response_template TEXT NOT NULL DEFAULT '',
  response_type TEXT NOT NULL DEFAULT 'reply',
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  enabled_when_offline BOOLEAN NOT NULL DEFAULT TRUE,
  enabled_when_online BOOLEAN NOT NULL DEFAULT TRUE,
  updated_by TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS custom_commands_platform_name_idx
  ON custom_commands (platform, lower(name));
