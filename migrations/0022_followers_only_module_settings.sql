CREATE TABLE IF NOT EXISTS followers_only_module_settings (
  id SMALLINT PRIMARY KEY CHECK (id = 1),
  enabled BOOLEAN NOT NULL DEFAULT FALSE,
  auto_disable_after_minutes INTEGER NOT NULL DEFAULT 30,
  updated_by TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO followers_only_module_settings (
  id,
  enabled,
  auto_disable_after_minutes,
  updated_by
)
VALUES (1, FALSE, 30, '')
ON CONFLICT (id) DO NOTHING;
