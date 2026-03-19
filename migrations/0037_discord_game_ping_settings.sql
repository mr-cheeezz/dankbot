ALTER TABLE discord_bot_settings
    ADD COLUMN IF NOT EXISTS game_ping_json JSONB NOT NULL DEFAULT '{}'::jsonb;
