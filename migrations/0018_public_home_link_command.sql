ALTER TABLE public_home_settings
	ADD COLUMN IF NOT EXISTS roblox_link_command_target TEXT NOT NULL DEFAULT 'dankbot',
	ADD COLUMN IF NOT EXISTS roblox_link_command_template TEXT NOT NULL DEFAULT '';
