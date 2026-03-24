CREATE TABLE IF NOT EXISTS alert_settings (
  id SMALLINT PRIMARY KEY,
  entries_json JSONB NOT NULL DEFAULT '[]'::jsonb,
  updated_by TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO alert_settings (
  id,
  entries_json,
  updated_by,
  created_at,
  updated_at
)
VALUES (1, '[]'::jsonb, '', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

