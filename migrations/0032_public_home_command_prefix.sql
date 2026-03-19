ALTER TABLE public_home_settings
	ADD COLUMN IF NOT EXISTS command_prefix TEXT NOT NULL DEFAULT '!';

UPDATE public_home_settings
SET command_prefix = '!'
WHERE TRIM(command_prefix) = '';
