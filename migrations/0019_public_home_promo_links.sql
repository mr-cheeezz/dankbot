ALTER TABLE public_home_settings
	ADD COLUMN IF NOT EXISTS promo_links_json TEXT NOT NULL DEFAULT '[]';
