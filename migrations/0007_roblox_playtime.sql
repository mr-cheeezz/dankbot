CREATE TABLE IF NOT EXISTS roblox_game_playtime (
    universe_id BIGINT PRIMARY KEY,
    root_place_id BIGINT NOT NULL DEFAULT 0,
    game_name TEXT NOT NULL DEFAULT '',
    total_seconds BIGINT NOT NULL DEFAULT 0,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_roblox_game_playtime_total_seconds
    ON roblox_game_playtime (total_seconds DESC, updated_at DESC);
