ALTER TABLE website_overlay_settings
  ADD COLUMN IF NOT EXISTS poll_scale DOUBLE PRECISION NOT NULL DEFAULT 1,
  ADD COLUMN IF NOT EXISTS poll_bar_color TEXT NOT NULL DEFAULT '#7dd3fc',
  ADD COLUMN IF NOT EXISTS poll_title_color TEXT NOT NULL DEFAULT '#ffffff',
  ADD COLUMN IF NOT EXISTS poll_text_color TEXT NOT NULL DEFAULT '#f8fafc',
  ADD COLUMN IF NOT EXISTS prediction_scale DOUBLE PRECISION NOT NULL DEFAULT 1,
  ADD COLUMN IF NOT EXISTS prediction_left_color TEXT NOT NULL DEFAULT '#60a5fa',
  ADD COLUMN IF NOT EXISTS prediction_right_color TEXT NOT NULL DEFAULT '#f472b6',
  ADD COLUMN IF NOT EXISTS prediction_text_color TEXT NOT NULL DEFAULT '#ffffff',
  ADD COLUMN IF NOT EXISTS prediction_track_color TEXT NOT NULL DEFAULT 'rgba(15, 23, 42, 0.28)';

UPDATE website_overlay_settings
SET
  prediction_position = 'top-center'
WHERE id = 1;
