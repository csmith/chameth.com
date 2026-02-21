package db

import (
	"context"
	"fmt"
	"time"

	"chameth.com/chameth.com/metrics"
)

func GetFilmReviewByID(ctx context.Context, id int) (*FilmReview, error) {
	metrics.LogQuery(ctx)
	var review FilmReview
	err := db.GetContext(ctx, &review, "SELECT id, film_id, watched_date, rating, is_rewatch, has_spoilers, review_text, published FROM film_reviews WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func GetFilmReviewsByFilmID(ctx context.Context, filmID int) ([]FilmReview, error) {
	metrics.LogQuery(ctx)
	var reviews []FilmReview
	err := db.SelectContext(ctx, &reviews, "SELECT id, film_id, watched_date, rating, is_rewatch, has_spoilers, review_text, published FROM film_reviews WHERE film_id = $1 ORDER BY watched_date DESC", filmID)
	if err != nil {
		return nil, err
	}
	return reviews, nil
}

func CreateFilmReview(ctx context.Context, filmID int, rating int, watchedDate time.Time, isRewatch, hasSpoilers, published bool, reviewText string) (int, error) {
	metrics.LogQuery(ctx)
	var id int
	err := db.QueryRowContext(ctx, `
		INSERT INTO film_reviews (film_id, rating, watched_date, is_rewatch, has_spoilers, review_text, published)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, filmID, rating, watchedDate, isRewatch, hasSpoilers, reviewText, published).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create film review: %w", err)
	}
	return id, nil
}

func UpdateFilmReview(ctx context.Context, id int, rating int, watchedDate string, isRewatch, hasSpoilers, published bool, reviewText string) error {
	metrics.LogQuery(ctx)
	_, err := db.ExecContext(ctx, `
		UPDATE film_reviews
		SET rating = $1, watched_date = $2, is_rewatch = $3, has_spoilers = $4, review_text = $5, published = $6
		WHERE id = $7
	`, rating, watchedDate, isRewatch, hasSpoilers, reviewText, published, id)
	if err != nil {
		return fmt.Errorf("failed to update film review: %w", err)
	}
	return nil
}

func GetFilmReviewWithFilmAndPoster(ctx context.Context, reviewID int) (*FilmReviewWithFilmAndPoster, error) {
	metrics.LogQuery(ctx)
	query := `
		SELECT
			fr.id as "filmreview.id", fr.film_id as "filmreview.film_id", fr.watched_date as "filmreview.watched_date",
			fr.rating as "filmreview.rating", fr.is_rewatch as "filmreview.is_rewatch", fr.has_spoilers as "filmreview.has_spoilers",
			fr.review_text as "filmreview.review_text", fr.published as "filmreview.published",
			f.id as "film.id", f.tmdb_id as "film.tmdb_id", f.title as "film.title", f.year as "film.year",
			f.overview as "film.overview", f.runtime as "film.runtime", f.published as "film.published",
			f.path as "film.path",
			mr.path as "poster.path", mr.media_id as "poster.media_id", mr.description as "poster.description",
			mr.caption as "poster.caption", mr.role as "poster.role", mr.entity_type as "poster.entity_type",
			mr.entity_id as "poster.entity_id",
			m.id as "poster.id", m.content_type as "poster.content_type", m.original_filename as "poster.original_filename",
			m.width as "poster.width", m.height as "poster.height", m.parent_media_id as "poster.parent_media_id"
		FROM film_reviews fr
		JOIN films f ON fr.film_id = f.id
		JOIN media_relations mr ON mr.entity_type = 'film' AND mr.entity_id = f.id AND mr.role = 'poster'
		JOIN media m ON mr.media_id = m.id
		WHERE fr.id = $1
	`

	var result FilmReviewWithFilmAndPoster
	err := db.GetContext(ctx, &result, query, reviewID)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func GetAllPublishedFilmReviewsWithFilmAndPosters(ctx context.Context) ([]FilmReviewWithFilmAndPoster, error) {
	metrics.LogQuery(ctx)
	query := `
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
		WHERE fr.published = true
		ORDER BY fr.watched_date DESC
	`

	var results []FilmReviewWithFilmAndPoster
	err := db.SelectContext(ctx, &results, query)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func GetRecentPublishedFilmReviewsWithFilmAndPosters(ctx context.Context, limit int) ([]FilmReviewWithFilmAndPoster, error) {
	metrics.LogQuery(ctx)
	query := `
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
		LEFT JOIN media_relations mr ON mr.entity_type = 'film' AND mr.entity_id = f.id AND mr.role = 'poster'
		LEFT JOIN media m ON mr.media_id = m.id
		WHERE fr.published = true
		ORDER BY fr.watched_date DESC
		LIMIT $1
	`

	var results []FilmReviewWithFilmAndPoster
	err := db.SelectContext(ctx, &results, query, limit)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func GetPublishedFilmReviewsWithFilmAndPostersByDateRange(ctx context.Context, startDate, endDate time.Time) ([]FilmReviewWithFilmAndPoster, error) {
	metrics.LogQuery(ctx)
	query := `
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
	`

	var results []FilmReviewWithFilmAndPoster
	err := db.SelectContext(ctx, &results, query, startDate, endDate)
	if err != nil {
		return nil, err
	}

	return results, nil
}
