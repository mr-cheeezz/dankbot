CREATE TABLE IF NOT EXISTS discord_log_settings (
  id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
  enabled BOOLEAN NOT NULL DEFAULT false,
  channel_id TEXT NOT NULL DEFAULT '',
  log_chat_messages BOOLEAN NOT NULL DEFAULT false,
  log_mod_actions BOOLEAN NOT NULL DEFAULT true,
  log_audit_logs BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO discord_log_settings (
  id,
  enabled,
  channel_id,
  log_chat_messages,
  log_mod_actions,
  log_audit_logs,
  created_at,
  updated_at
)
VALUES (
  1,
  false,
  '',
  false,
  true,
  true,
  NOW(),
  NOW()
)
ON CONFLICT (id) DO NOTHING;

