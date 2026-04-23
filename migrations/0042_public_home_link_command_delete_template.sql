ALTER TABLE public_home_settings
ADD COLUMN IF NOT EXISTS roblox_link_command_delete_template TEXT NOT NULL DEFAULT '';
