CREATE TABLE IF NOT EXISTS tabs_module_settings (
  id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
  enabled BOOLEAN NOT NULL DEFAULT true,
  interest_rate_percent NUMERIC(8,4) NOT NULL DEFAULT 0.0000,
  interest_every_days INTEGER NOT NULL DEFAULT 7,
  updated_by TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO tabs_module_settings (
  id,
  enabled,
  interest_rate_percent,
  interest_every_days,
  updated_by,
  created_at,
  updated_at
)
VALUES (
  1,
  true,
  0.0000,
  7,
  '',
  NOW(),
  NOW()
)
ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS user_tabs (
  login TEXT PRIMARY KEY,
  display_name TEXT NOT NULL DEFAULT '',
  balance_cents BIGINT NOT NULL DEFAULT 0,
  last_interest_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_tabs_updated_at
  ON user_tabs (updated_at DESC);

INSERT INTO module_catalog (
  id,
  display_name,
  state,
  detail,
  commands,
  settings_schema,
  sort_order
)
VALUES (
  'tabs',
  'Tabs',
  'finance',
  'Track viewer tabs with optional automatic interest after a configurable number of days.',
  '["!tab <user>", "!tab add <user> <amount>", "!tab set <user> <amount>", "!tab paid <user>"]'::jsonb,
  '[
    {
      "id": "enabled",
      "label": "Enabled",
      "type": "boolean",
      "helper_text": "Turn tab tracking commands on/off."
    },
    {
      "id": "interest-rate-percent",
      "label": "Interest rate (%)",
      "type": "number",
      "helper_text": "Percent added each interval when a tab remains unpaid."
    },
    {
      "id": "interest-every-days",
      "label": "Interest interval (days)",
      "type": "number",
      "helper_text": "How many days must pass before another interest charge is applied."
    }
  ]'::jsonb,
  60
)
ON CONFLICT (id) DO UPDATE
SET
  display_name = EXCLUDED.display_name,
  state = EXCLUDED.state,
  detail = EXCLUDED.detail,
  commands = EXCLUDED.commands,
  settings_schema = EXCLUDED.settings_schema,
  sort_order = EXCLUDED.sort_order,
  updated_at = NOW();

