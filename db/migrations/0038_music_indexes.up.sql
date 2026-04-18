CREATE INDEX idx_music_plays_track_id ON music_plays(track_id);
CREATE INDEX idx_music_plays_played_at ON music_plays(played_at DESC);
CREATE INDEX idx_music_tracks_album_id ON music_tracks(album_id);
CREATE INDEX idx_music_albums_artist_id ON music_albums(artist_id);
CREATE INDEX idx_media_relations_entity ON media_relations(entity_type, entity_id, role);
