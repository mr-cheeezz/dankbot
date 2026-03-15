CREATE TABLE IF NOT EXISTS blocked_terms (
  id TEXT PRIMARY KEY,
  pattern TEXT NOT NULL,
  is_regex BOOLEAN NOT NULL DEFAULT FALSE,
  action TEXT NOT NULL,
  timeout_seconds INTEGER NOT NULL DEFAULT 0,
  reason TEXT NOT NULL DEFAULT '',
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS blocked_terms_enabled_idx
  ON blocked_terms (enabled, pattern);
