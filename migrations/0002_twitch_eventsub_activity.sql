CREATE TABLE IF NOT EXISTS twitch_poll_events (
    id BIGSERIAL PRIMARY KEY,
    twitch_subscription_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    poll_id TEXT NOT NULL,
    broadcaster_user_id TEXT NOT NULL,
    broadcaster_user_login TEXT NOT NULL DEFAULT '',
    broadcaster_user_name TEXT NOT NULL DEFAULT '',
    title TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT '',
    started_at TIMESTAMPTZ NULL,
    ended_at TIMESTAMPTZ NULL,
    raw_event JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_twitch_poll_events_poll_id
    ON twitch_poll_events (poll_id, created_at DESC);

CREATE TABLE IF NOT EXISTS twitch_poll_event_choices (
    id BIGSERIAL PRIMARY KEY,
    poll_event_id BIGINT NOT NULL REFERENCES twitch_poll_events (id) ON DELETE CASCADE,
    choice_id TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    votes INTEGER NOT NULL DEFAULT 0,
    channel_points_votes INTEGER NOT NULL DEFAULT 0,
    bits_votes INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_twitch_poll_event_choices_poll_event_id
    ON twitch_poll_event_choices (poll_event_id);

CREATE TABLE IF NOT EXISTS twitch_channel_point_redemptions (
    redemption_id TEXT PRIMARY KEY,
    twitch_subscription_id TEXT NOT NULL,
    broadcaster_user_id TEXT NOT NULL,
    broadcaster_user_login TEXT NOT NULL DEFAULT '',
    broadcaster_user_name TEXT NOT NULL DEFAULT '',
    user_id TEXT NOT NULL,
    user_login TEXT NOT NULL DEFAULT '',
    user_name TEXT NOT NULL DEFAULT '',
    user_input TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT '',
    redeemed_at TIMESTAMPTZ NULL,
    reward_id TEXT NOT NULL,
    reward_title TEXT NOT NULL DEFAULT '',
    reward_cost INTEGER NOT NULL DEFAULT 0,
    reward_prompt TEXT NOT NULL DEFAULT '',
    raw_event JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
