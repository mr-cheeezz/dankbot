CREATE TABLE IF NOT EXISTS website_overlay_settings (
  id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
  polls_enabled BOOLEAN NOT NULL DEFAULT true,
  predictions_enabled BOOLEAN NOT NULL DEFAULT true,
  poll_position TEXT NOT NULL DEFAULT 'bottom-left',
  prediction_position TEXT NOT NULL DEFAULT 'bottom-right',
  poll_offset_x INTEGER NOT NULL DEFAULT 24,
  poll_offset_y INTEGER NOT NULL DEFAULT 24,
  prediction_offset_x INTEGER NOT NULL DEFAULT 24,
  prediction_offset_y INTEGER NOT NULL DEFAULT 24,
  updated_by TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO website_overlay_settings (
  id,
  polls_enabled,
  predictions_enabled,
  poll_position,
  prediction_position,
  poll_offset_x,
  poll_offset_y,
  prediction_offset_x,
  prediction_offset_y,
  updated_by,
  created_at,
  updated_at
)
VALUES (
  1,
  true,
  true,
  'bottom-left',
  'bottom-right',
  24,
  24,
  24,
  24,
  '',
  NOW(),
  NOW()
)
ON CONFLICT (id) DO NOTHING;
