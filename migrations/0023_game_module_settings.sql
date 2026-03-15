CREATE TABLE IF NOT EXISTS game_module_settings (
  id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
  keyword_response TEXT NOT NULL DEFAULT '@{target}, use !game to see what {streamer} is currently playing.',
  updated_by TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO game_module_settings (
  id,
  keyword_response,
  updated_by,
  created_at,
  updated_at
)
VALUES (
  1,
  '@{target}, use !game to see what {streamer} is currently playing.',
  '',
  NOW(),
  NOW()
)
ON CONFLICT (id) DO NOTHING;
