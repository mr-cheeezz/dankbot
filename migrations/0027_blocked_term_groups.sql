ALTER TABLE blocked_terms
  ADD COLUMN IF NOT EXISTS name TEXT NOT NULL DEFAULT '';

ALTER TABLE blocked_terms
  ADD COLUMN IF NOT EXISTS phrase_groups JSONB NOT NULL DEFAULT '[]'::jsonb;

UPDATE blocked_terms
SET name = pattern
WHERE COALESCE(name, '') = ''
  AND COALESCE(pattern, '') <> '';

CREATE INDEX IF NOT EXISTS blocked_terms_name_idx
  ON blocked_terms (name);
