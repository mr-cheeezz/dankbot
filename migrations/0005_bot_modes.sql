CREATE TABLE IF NOT EXISTS bot_modes (
    mode_key TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    is_builtin BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS bot_state (
    state_key TEXT PRIMARY KEY,
    current_mode_key TEXT NOT NULL REFERENCES bot_modes (mode_key),
    current_mode_param TEXT NOT NULL DEFAULT '',
    killswitch_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    updated_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
