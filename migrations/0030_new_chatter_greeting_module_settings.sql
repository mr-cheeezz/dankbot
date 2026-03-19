CREATE TABLE IF NOT EXISTS new_chatter_greeting_module_settings (
  id INT PRIMARY KEY,
  enabled BOOLEAN NOT NULL DEFAULT FALSE,
  messages_text TEXT NOT NULL DEFAULT '',
  updated_by TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO new_chatter_greeting_module_settings (
  id,
  enabled,
  messages_text,
  updated_by
)
VALUES (
  1,
  FALSE,
  'Welcome to chat, {user}!' || E'\n' || 'Glad you''re here, {display_name}!',
  ''
)
ON CONFLICT (id) DO NOTHING;

