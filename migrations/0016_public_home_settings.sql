CREATE TABLE IF NOT EXISTS public_home_settings (
	id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
	show_now_playing BOOLEAN NOT NULL DEFAULT TRUE,
	show_now_playing_album_art BOOLEAN NOT NULL DEFAULT TRUE,
	show_now_playing_progress BOOLEAN NOT NULL DEFAULT TRUE,
	show_now_playing_links BOOLEAN NOT NULL DEFAULT TRUE,
	updated_by TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
