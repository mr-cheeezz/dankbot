ALTER TABLE now_playing_module_settings
  ADD COLUMN IF NOT EXISTS song_change_message_template TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS song_command_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN IF NOT EXISTS song_next_command_enabled BOOLEAN NOT NULL DEFAULT TRUE,
  ADD COLUMN IF NOT EXISTS song_last_command_enabled BOOLEAN NOT NULL DEFAULT TRUE;

UPDATE module_catalog
SET
  settings_schema = '[
    {
      "id": "viewer-question-ai-detection",
      "label": "Use AI intent detection",
      "type": "boolean",
      "helper_text": "Helps avoid false positives when chat mentions songs without asking."
    },
    {
      "id": "viewer-question-response",
      "label": "Viewer question response",
      "type": "textarea",
      "helper_text": "Supports @{target}, {target}, and {streamer} placeholders."
    },
    {
      "id": "song-change-message-template",
      "label": "Song change message",
      "type": "textarea",
      "helper_text": "Chat message sent when track changes. Supports {streamer}, {song}, and {track}."
    },
    {
      "id": "song-command-enabled",
      "label": "Enable !song",
      "type": "boolean",
      "helper_text": "Enable/disable the !song command that shows the current song."
    },
    {
      "id": "song-next-command-enabled",
      "label": "Enable !song next",
      "type": "boolean",
      "helper_text": "Enable/disable the !song next command."
    },
    {
      "id": "song-last-command-enabled",
      "label": "Enable !song last",
      "type": "boolean",
      "helper_text": "Enable/disable the !song last command."
    }
  ]'::jsonb,
  updated_at = NOW()
WHERE id = 'now-playing';
