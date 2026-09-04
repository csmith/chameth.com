-- Music data now comes straight from the Bonus Metal API via the cached
-- data shortcodes, so the local Navidrome mirror is no longer needed. The
-- rehosted album and artist art stays: it is re-keyed from the local music
-- ids to the service's stable subsonic ids, so the shortcodes keep serving
-- the existing images in place. Subsonic ids are alphanumeric, so these
-- paths cannot collide with the integer-keyed ones they replace. The
-- media_relations update trigger keeps the paths table in sync.

UPDATE media_relations mr
SET path = '/music/albums/' || a.subsonic_id || '/cover.jpg'
FROM music_albums a
WHERE mr.entity_type = 'album'
  AND mr.entity_id = a.id
  AND mr.role = 'image';

UPDATE media_relations mr
SET path = '/music/artists/' || ar.subsonic_id || '/cover.jpg'
FROM music_artists ar
WHERE mr.entity_type = 'artist'
  AND mr.entity_id = ar.id
  AND mr.role = 'image';

DROP TABLE music_plays;
DROP TABLE music_tracks;
DROP TABLE music_albums;
DROP TABLE music_artists;
