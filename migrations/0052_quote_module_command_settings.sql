ALTER TABLE quote_module_settings
  ADD COLUMN IF NOT EXISTS quote_command_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN IF NOT EXISTS add_command_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN IF NOT EXISTS edit_command_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN IF NOT EXISTS delete_command_enabled BOOLEAN NOT NULL DEFAULT TRUE;

UPDATE module_catalog
SET
  detail = 'Website-managed quote storage with configurable chat commands and a saved quote library.',
  commands = '["!quote", "!add quote", "!create quote", "!edit quote", "!del quote", "!rm quote"]'::jsonb,
  settings_schema = '[
    {
      "id": "quote-command-enabled",
      "label": "Enable !quote lookup",
      "type": "boolean",
      "helper_text": "Lets chat pull a random quote or a specific quote number."
    },
    {
      "id": "add-command-enabled",
      "label": "Enable add quote commands",
      "type": "boolean",
      "helper_text": "Controls !add quote and !create quote."
    },
    {
      "id": "edit-command-enabled",
      "label": "Enable edit quote command",
      "type": "boolean",
      "helper_text": "Controls !edit quote for moderators and the broadcaster."
    },
    {
      "id": "delete-command-enabled",
      "label": "Enable delete quote commands",
      "type": "boolean",
      "helper_text": "Controls !del quote and !rm quote."
    }
  ]'::jsonb,
  updated_at = NOW()
WHERE id = 'quotes';
