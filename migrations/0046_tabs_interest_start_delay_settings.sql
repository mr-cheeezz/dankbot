ALTER TABLE tabs_module_settings
  ADD COLUMN IF NOT EXISTS interest_start_delay_mode TEXT NOT NULL DEFAULT 'week',
  ADD COLUMN IF NOT EXISTS interest_start_delay_value INTEGER NOT NULL DEFAULT 1,
  ADD COLUMN IF NOT EXISTS interest_start_delay_unit TEXT NOT NULL DEFAULT 'weeks';

UPDATE module_catalog
SET
  settings_schema = '[
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
    },
    {
      "id": "interest-start-delay-mode",
      "label": "Interest start delay",
      "type": "select",
      "helper_text": "When interest should begin: day, week, month, or custom.",
      "options": ["day", "week", "month", "custom"]
    },
    {
      "id": "interest-start-delay-value",
      "label": "Custom delay amount",
      "type": "number",
      "helper_text": "Used only when start delay is custom."
    },
    {
      "id": "interest-start-delay-unit",
      "label": "Custom delay unit",
      "type": "select",
      "helper_text": "Used only when start delay is custom.",
      "options": ["days", "weeks", "months"]
    }
  ]'::jsonb,
  updated_at = NOW()
WHERE id = 'tabs';
