CREATE TABLE IF NOT EXISTS now_playing_module_settings (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    keyword_response TEXT NOT NULL DEFAULT '',
    updated_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
