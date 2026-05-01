ALTER TABLE tabs_module_settings
  ADD COLUMN IF NOT EXISTS interest_interval_mode TEXT NOT NULL DEFAULT 'weekly',
  ADD COLUMN IF NOT EXISTS interest_interval_custom_days INTEGER NOT NULL DEFAULT 7,
  ADD COLUMN IF NOT EXISTS grace_period_days INTEGER NOT NULL DEFAULT 7;

UPDATE tabs_module_settings
SET
  interest_interval_mode = CASE
    WHEN interest_every_days <= 1 THEN 'daily'
    WHEN interest_every_days = 2 THEN 'bi-daily'
    WHEN interest_every_days >= 30 THEN 'monthly'
    WHEN interest_every_days >= 14 THEN 'bi-weekly'
    ELSE 'weekly'
  END,
  interest_interval_custom_days = LEAST(GREATEST(interest_every_days, 1), 30),
  grace_period_days = LEAST(GREATEST(
    CASE
      WHEN interest_start_delay_mode = 'day' THEN 1
      WHEN interest_start_delay_mode = 'week' THEN 7
      WHEN interest_start_delay_mode = 'month' THEN 30
      ELSE interest_start_delay_value *
        CASE
          WHEN interest_start_delay_unit = 'weeks' THEN 7
          WHEN interest_start_delay_unit = 'months' THEN 30
          ELSE 1
        END
    END
  , 1), 30)
WHERE id = 1;

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
      "helper_text": "Percent added each interval while a tab is unpaid."
    },
    {
      "id": "interest-interval",
      "label": "Interest compounding",
      "type": "select",
      "helper_text": "How often interest compounds.",
      "options": ["daily", "bi-daily", "weekly", "bi-weekly", "monthly", "custom"]
    },
    {
      "id": "interest-interval-custom-days",
      "label": "Custom interval (days)",
      "type": "number",
      "helper_text": "Only used when compounding is custom. Max 30 days."
    },
    {
      "id": "grace-period-days",
      "label": "Grace period (days)",
      "type": "number",
      "helper_text": "Interest grace after a tab is paid. Only tab paid resets this grace."
    }
  ]'::jsonb,
  updated_at = NOW()
WHERE id = 'tabs';

