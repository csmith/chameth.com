package watchedfilms

import (
	"context"
	"time"

	"chameth.com/chameth.com/db"
)

func query(ctx context.Context, startDate, endDate time.Time) ([]db.FilmReviewWithFilmAndPoster, error) {
	return db.Select[db.FilmReviewWithFilmAndPoster](ctx, `
		SELECT
			fr.id as "filmreview.id", fr.film_id as "filmreview.film_id", fr.watched_date as "filmreview.watched_date",
			fr.rating as "filmreview.rating", fr.is_rewatch as "filmreview.is_rewatch", fr.has_spoilers as "filmreview.has_spoilers",
			fr.review_text as "filmreview.review_text", fr.published as "filmreview.published",
			f.id as "film.id", f.tmdb_id as "film.tmdb_id", f.title as "film.title", f.year as "film.year",
			f.overview as "film.overview", f.runtime as "film.runtime", f.published as "film.published", f.path as "film.path",
			mr.path as "poster.path", mr.media_id as "poster.media_id", mr.description as "poster.description",
			mr.caption as "poster.caption", mr.role as "poster.role", mr.entity_type as "poster.entity_type",
			mr.entity_id as "poster.entity_id",
			m.id as "poster.id", m.content_type as "poster.content_type", m.original_filename as "poster.original_filename",
			m.width as "poster.width", m.height as "poster.height", m.parent_media_id as "poster.parent_media_id"
		FROM film_reviews fr
		JOIN films f ON fr.film_id = f.id
		JOIN media_relations mr ON mr.entity_type = 'film' AND mr.entity_id = f.id AND mr.role = 'poster'
		JOIN media m ON mr.media_id = m.id
		WHERE fr.published = true AND fr.watched_date >= $1 AND fr.watched_date <= $2
		ORDER BY fr.watched_date ASC
	`, startDate, endDate)
}
