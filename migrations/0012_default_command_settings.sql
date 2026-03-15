CREATE TABLE IF NOT EXISTS default_command_settings (
	command_name TEXT PRIMARY KEY,
	enabled BOOLEAN NOT NULL DEFAULT TRUE,
	config_json JSONB NOT NULL DEFAULT '{}'::jsonb,
	updated_by TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
