CREATE TABLE IF NOT EXISTS discord_bot_installation (
  id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
  guild_id TEXT NOT NULL DEFAULT '',
  installer_user_id TEXT NOT NULL DEFAULT '',
  installer_username TEXT NOT NULL DEFAULT '',
  permissions TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
