ALTER TABLE followers_only_module_settings
  ADD COLUMN IF NOT EXISTS enabled_when_offline BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE module_catalog
SET
  settings_schema = '[
    {
      "id": "enabled-when-offline",
      "label": "Enabled when stream offline",
      "type": "boolean",
      "helper_text": "Keep auto followers-only checks running even while the stream is offline."
    },
    {
      "id": "auto-disable-minutes",
      "label": "Auto-disable after",
      "type": "number",
      "helper_text": "How many minutes followers-only can stay enabled before DankBot turns it back off."
    }
  ]'::jsonb,
  updated_at = NOW()
WHERE id = 'auto-followers-only';
