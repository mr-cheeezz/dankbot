CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    platform TEXT NOT NULL DEFAULT '',
    actor_id TEXT NOT NULL DEFAULT '',
    actor_name TEXT NOT NULL DEFAULT '',
    command_name TEXT NOT NULL,
    detail TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS audit_logs_created_at_idx
    ON audit_logs (created_at DESC);
