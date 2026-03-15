CREATE TABLE IF NOT EXISTS discord_bot_settings (
    id SMALLINT PRIMARY KEY CHECK (id = 1),
    guild_id TEXT NOT NULL DEFAULT '',
    default_channel_id TEXT NOT NULL DEFAULT '',
    ping_roles_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
