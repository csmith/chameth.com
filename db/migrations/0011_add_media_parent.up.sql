ALTER TABLE media
    ADD COLUMN parent_media_id INT REFERENCES media (id);

UPDATE media
SET parent_media_id = (SELECT parent_media.id
                       FROM media_relations AS variant_rel
                                JOIN media_relations AS parent_rel
                                     ON regexp_replace(variant_rel.slug, '\.(webp|avif)$', '') =
                                        regexp_replace(parent_rel.slug, '\.(jpg|jpeg|png)$', '')
                                         AND parent_rel.slug ~ '\.(jpg|jpeg|png)$'
                                JOIN media AS parent_media ON parent_rel.media_id = parent_media.id
                       WHERE variant_rel.media_id = media.id
                       LIMIT 1)
WHERE original_filename ~ '\.(webp|avif)$';
