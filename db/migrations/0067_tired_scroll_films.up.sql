-- Films are now managed in Tired Scroll. The site keeps a mirror of its
-- public index (films, lists and list entries) for routing, list pages,
-- search and the sitemap, refreshed by a background sync; reviews and film
-- details are read through the cached data shortcodes. Tired Scroll kept the
-- ids from these tables, so existing rows (and poster paths) carry over.

-- Letterboxd links now come from Tired Scroll.
DELETE FROM syndications
WHERE name = 'Letterboxd'
  AND path IN (SELECT path FROM films UNION SELECT path FROM film_lists);

DROP TABLE film_reviews;

ALTER TABLE films
    ALTER COLUMN id DROP IDENTITY,
    ALTER COLUMN path SET NOT NULL,
    DROP COLUMN tmdb_id,
    DROP COLUMN overview,
    DROP COLUMN runtime,
    DROP COLUMN published,
    ADD COLUMN watch_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN average_rating DOUBLE PRECISION,
    ADD COLUMN last_watched DATE,
    ADD COLUMN poster_updated_at TIMESTAMPTZ;

ALTER TABLE film_lists
    ALTER COLUMN id DROP IDENTITY,
    DROP COLUMN published,
    ADD COLUMN letterboxd_url VARCHAR NOT NULL DEFAULT '',
    ADD COLUMN ranking BOOLEAN NOT NULL DEFAULT false;

UPDATE film_lists SET description = '' WHERE description IS NULL;
ALTER TABLE film_lists
    ALTER COLUMN description SET NOT NULL,
    ALTER COLUMN path SET NOT NULL;
