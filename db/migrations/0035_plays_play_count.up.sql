TRUNCATE TABLE music_plays;
TRUNCATE TABLE unmatched_music_plays;

ALTER TABLE music_plays DROP COLUMN play_id;
ALTER TABLE music_plays ADD COLUMN play_count INTEGER NOT NULL DEFAULT 1;

ALTER TABLE unmatched_music_plays DROP COLUMN play_id;
ALTER TABLE unmatched_music_plays ADD COLUMN play_count INTEGER NOT NULL DEFAULT 1;
