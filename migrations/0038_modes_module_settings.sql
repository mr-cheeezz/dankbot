CREATE TABLE IF NOT EXISTS modes_module_settings (
  id SMALLINT PRIMARY KEY,
  legacy_commands_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  updated_by TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO modes_module_settings (
  id,
  legacy_commands_enabled,
  updated_by,
  created_at,
  updated_at
)
VALUES (1, FALSE, '', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

