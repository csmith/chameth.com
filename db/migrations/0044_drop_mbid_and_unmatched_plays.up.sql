ALTER TABLE music_artists DROP COLUMN music_brainz_id;
ALTER TABLE music_albums DROP COLUMN music_brainz_id;
ALTER TABLE music_tracks DROP COLUMN music_brainz_id;

DROP TABLE unmatched_music_plays;
