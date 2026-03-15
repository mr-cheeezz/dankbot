CREATE TABLE IF NOT EXISTS keywords (
	id BIGSERIAL PRIMARY KEY,
	trigger TEXT NOT NULL,
	response TEXT NOT NULL,
	created_by TEXT NOT NULL DEFAULT '',
	updated_by TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS keywords_trigger_unique_idx
ON keywords (LOWER(trigger));
