-- Board game plays now come straight from the Magic Meters API via the
-- cached data shortcodes, so the local play log is no longer needed.
-- Magic Meters serves JPEG art keyed by BGG id, which is what the existing
-- image mapping already uses, so those images are untouched; the non-JPEG
-- copies are dropped, as nothing will ever serve them again.
DELETE FROM media_relations mr
USING media m
WHERE mr.media_id = m.id
  AND mr.entity_type = 'boardgame'
  AND m.content_type <> 'image/jpeg';

DELETE FROM media m
WHERE m.original_filename LIKE 'boardgame-%'
  AND NOT EXISTS (SELECT 1 FROM media_relations mr WHERE mr.media_id = m.id);

DROP TABLE boardgame_plays;
DROP TABLE boardgame_games;
