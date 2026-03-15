ALTER TABLE bot_modes
    ADD COLUMN IF NOT EXISTS timer_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS timer_message TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS timer_interval_seconds INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS last_timer_sent_at TIMESTAMPTZ NULL;

CREATE TABLE IF NOT EXISTS bot_social_promotions (
    id BIGSERIAL PRIMARY KEY,
    command_text TEXT NOT NULL,
    interval_seconds INTEGER NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_sent_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_bot_social_promotions_enabled
    ON bot_social_promotions (enabled, id);
