CREATE TABLE IF NOT EXISTS twitch_accounts (
    kind TEXT PRIMARY KEY,
    twitch_user_id TEXT NOT NULL,
    login TEXT NOT NULL,
    display_name TEXT NOT NULL DEFAULT '',
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL DEFAULT '',
    scopes JSONB NOT NULL DEFAULT '[]'::jsonb,
    token_type TEXT NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ NULL,
    last_validated_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS twitch_eventsub_subscriptions (
    twitch_subscription_id TEXT PRIMARY KEY,
    subscription_type TEXT NOT NULL,
    subscription_version TEXT NOT NULL,
    status TEXT NOT NULL,
    condition JSONB NOT NULL DEFAULT '{}'::jsonb,
    callback_url TEXT NOT NULL,
    transport_method TEXT NOT NULL DEFAULT 'webhook',
    secret_fingerprint TEXT NOT NULL DEFAULT '',
    secret_version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_notification_at TIMESTAMPTZ NULL,
    last_revoked_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_twitch_eventsub_subscriptions_callback_url
    ON twitch_eventsub_subscriptions (callback_url);
