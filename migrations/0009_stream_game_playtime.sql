CREATE TABLE IF NOT EXISTS stream_game_playtime (
    game_key TEXT PRIMARY KEY,
    source TEXT NOT NULL,
    twitch_game_id TEXT NOT NULL DEFAULT '',
    roblox_universe_id BIGINT NOT NULL DEFAULT 0,
    game_name TEXT NOT NULL DEFAULT '',
    total_seconds BIGINT NOT NULL DEFAULT 0,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_stream_game_playtime_total_seconds
    ON stream_game_playtime (total_seconds DESC, updated_at DESC);

CREATE TABLE IF NOT EXISTS stream_game_playtime_sessions (
    id BIGSERIAL PRIMARY KEY,
    stream_session_id TEXT NOT NULL,
    game_key TEXT NOT NULL,
    source TEXT NOT NULL,
    twitch_game_id TEXT NOT NULL DEFAULT '',
    roblox_universe_id BIGINT NOT NULL DEFAULT 0,
    game_name TEXT NOT NULL DEFAULT '',
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ NOT NULL,
    duration_seconds BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_stream_game_playtime_sessions_stream_session_id
    ON stream_game_playtime_sessions (stream_session_id, ended_at DESC);

CREATE INDEX IF NOT EXISTS idx_stream_game_playtime_sessions_ended_at
    ON stream_game_playtime_sessions (ended_at DESC);
