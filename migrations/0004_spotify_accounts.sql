CREATE TABLE IF NOT EXISTS spotify_accounts (
    kind TEXT PRIMARY KEY,
    spotify_user_id TEXT NOT NULL,
    display_name TEXT NOT NULL DEFAULT '',
    email TEXT NOT NULL DEFAULT '',
    product TEXT NOT NULL DEFAULT '',
    country TEXT NOT NULL DEFAULT '',
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL DEFAULT '',
    scope TEXT NOT NULL DEFAULT '',
    token_type TEXT NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
