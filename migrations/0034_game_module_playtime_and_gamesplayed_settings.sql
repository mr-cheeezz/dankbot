-- Expand the existing "game" module settings to cover the !playtime and
-- !gamesplayed commands (these outputs used to be hardcoded).

ALTER TABLE game_module_settings
  ADD COLUMN IF NOT EXISTS playtime_template TEXT NOT NULL DEFAULT '{streamer} has been playing {game} for {duration}.',
  ADD COLUMN IF NOT EXISTS gamesplayed_template TEXT NOT NULL DEFAULT '{label}: {items}',
  ADD COLUMN IF NOT EXISTS gamesplayed_item_template TEXT NOT NULL DEFAULT '{game} ({duration})',
  ADD COLUMN IF NOT EXISTS gamesplayed_limit SMALLINT NOT NULL DEFAULT 5;

-- Make sure an existing row gets populated if it existed before these columns.
UPDATE game_module_settings
SET
  playtime_template = COALESCE(NULLIF(TRIM(playtime_template), ''), '{streamer} has been playing {game} for {duration}.'),
  gamesplayed_template = COALESCE(NULLIF(TRIM(gamesplayed_template), ''), '{label}: {items}'),
  gamesplayed_item_template = COALESCE(NULLIF(TRIM(gamesplayed_item_template), ''), '{game} ({duration})'),
  gamesplayed_limit = CASE
    WHEN gamesplayed_limit IS NULL OR gamesplayed_limit < 1 THEN 5
    ELSE gamesplayed_limit
  END
WHERE id = 1;

-- Update the module catalog schema so these fields appear in the module editor.
UPDATE module_catalog
SET
  settings_schema = '[
    {
      "id": "viewer-question-enabled",
      "label": "Viewer question keyword enabled",
      "type": "boolean",
      "helper_text": "Enable/disable the built-in what-game question keyword handling."
    },
    {
      "id": "viewer-question-ai-detection",
      "label": "Use AI intent detection",
      "type": "boolean",
      "helper_text": "Helps avoid false positives when people mention a game without asking."
    },
    {
      "id": "viewer-question-response",
      "label": "Viewer question response",
      "type": "textarea",
      "helper_text": "Supports @{target}, {target}, and {streamer} placeholders."
    },
    {
      "id": "playtime-template",
      "label": "!playtime response template",
      "type": "textarea",
      "helper_text": "Supports {streamer}, {game}, and {duration}."
    },
    {
      "id": "gamesplayed-template",
      "label": "!gamesplayed response template",
      "type": "textarea",
      "helper_text": "Supports {label} and {items}."
    },
    {
      "id": "gamesplayed-item-template",
      "label": "!gamesplayed item template",
      "type": "text",
      "helper_text": "Supports {game} and {duration}."
    },
    {
      "id": "gamesplayed-limit",
      "label": "Top games limit",
      "type": "number",
      "helper_text": "How many games to include in the !gamesplayed output."
    }
  ]'::jsonb,
  updated_at = NOW()
WHERE id = 'game';

