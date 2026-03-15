CREATE TABLE IF NOT EXISTS roblox_game_playtime_sessions (
    id BIGSERIAL PRIMARY KEY,
    stream_session_id TEXT NOT NULL,
    universe_id BIGINT NOT NULL,
    root_place_id BIGINT NOT NULL DEFAULT 0,
    game_name TEXT NOT NULL DEFAULT '',
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ NOT NULL,
    duration_seconds BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_roblox_game_playtime_sessions_stream_session_id
    ON roblox_game_playtime_sessions (stream_session_id, ended_at DESC);

CREATE INDEX IF NOT EXISTS idx_roblox_game_playtime_sessions_ended_at
    ON roblox_game_playtime_sessions (ended_at DESC);
