CREATE TABLE IF NOT EXISTS spam_filter_hype_settings (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  enabled BOOLEAN NOT NULL DEFAULT FALSE,
  disable_duration_seconds INTEGER NOT NULL DEFAULT 180,
  bits_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  bits_threshold INTEGER NOT NULL DEFAULT 1000,
  gifted_subs_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  gifted_subs_threshold INTEGER NOT NULL DEFAULT 10,
  raids_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  raids_threshold INTEGER NOT NULL DEFAULT 50,
  donations_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  donations_threshold NUMERIC(10,2) NOT NULL DEFAULT 25,
  disabled_filter_keys JSONB NOT NULL DEFAULT '[]'::jsonb,
  updated_by TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
