CREATE TABLE IF NOT EXISTS quote_module_settings (
    id SMALLINT PRIMARY KEY,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    updated_by TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO quote_module_settings (
    id,
    enabled,
    updated_by,
    created_at,
    updated_at
)
VALUES (1, TRUE, '', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;
