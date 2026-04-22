ALTER TABLE spam_filters
  ADD COLUMN IF NOT EXISTS repeat_offenders_enabled BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS repeat_multiplier DOUBLE PRECISION NOT NULL DEFAULT 1,
  ADD COLUMN IF NOT EXISTS repeat_memory_seconds INTEGER NOT NULL DEFAULT 600,
  ADD COLUMN IF NOT EXISTS repeat_until_stream_end BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE spam_filters
SET
  repeat_offenders_enabled = TRUE,
  repeat_multiplier = 2,
  repeat_memory_seconds = 600,
  repeat_until_stream_end = FALSE
WHERE filter_key = 'message-flood';
