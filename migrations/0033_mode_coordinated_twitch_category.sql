ALTER TABLE bot_modes
	ADD COLUMN IF NOT EXISTS coordinated_twitch_category_id TEXT NOT NULL DEFAULT '',
	ADD COLUMN IF NOT EXISTS coordinated_twitch_category_name TEXT NOT NULL DEFAULT '';
