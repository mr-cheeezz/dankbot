ALTER TABLE spam_filters
  ADD COLUMN IF NOT EXISTS impacted_roles JSONB NOT NULL DEFAULT '[]'::jsonb,
  ADD COLUMN IF NOT EXISTS excluded_roles JSONB NOT NULL DEFAULT '[]'::jsonb;
