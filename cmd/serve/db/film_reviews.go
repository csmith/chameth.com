package db

import (
	"fmt"
	"time"
)

func GetFilmReviewByID(id int) (*FilmReview, error) {
	var review FilmReview
	err := db.Get(&review, "SELECT id, film_id, watched_date, rating, is_rewatch, has_spoilers, review_text, published FROM film_reviews WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func GetFilmReviewsByFilmID(filmID int) ([]FilmReview, error) {
	var reviews []FilmReview
	err := db.Select(&reviews, "SELECT id, film_id, watched_date, rating, is_rewatch, has_spoilers, review_text, published FROM film_reviews WHERE film_id = $1 ORDER BY watched_date DESC", filmID)
	if err != nil {
		return nil, err
	}
	return reviews, nil
}

func GetLatestFilmReviewByFilmID(filmID int) (*FilmReview, error) {
	var review FilmReview
	err := db.Get(&review, "SELECT id, film_id, watched_date, rating, is_rewatch, has_spoilers, review_text, published FROM film_reviews WHERE film_id = $1 ORDER BY watched_date DESC LIMIT 1", filmID)
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func CreateFilmReview(filmID int, rating int, watchedDate time.Time, isRewatch, hasSpoilers, published bool, reviewText string) (int, error) {
	var id int
	err := db.QueryRow(`
		INSERT INTO film_reviews (film_id, rating, watched_date, is_rewatch, has_spoilers, review_text, published)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, filmID, rating, watchedDate, isRewatch, hasSpoilers, reviewText, published).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create film review: %w", err)
	}
	return id, nil
}

func UpdateFilmReview(id int, rating int, watchedDate string, isRewatch, hasSpoilers, published bool, reviewText string) error {
	_, err := db.Exec(`
		UPDATE film_reviews
		SET rating = $1, watched_date = $2, is_rewatch = $3, has_spoilers = $4, review_text = $5, published = $6
		WHERE id = $7
	`, rating, watchedDate, isRewatch, hasSpoilers, reviewText, published, id)
	if err != nil {
		return fmt.Errorf("failed to update film review: %w", err)
	}
	return nil
}

func GetFilmReviewWithFilmAndPoster(reviewID int) (*FilmReviewWithFilmAndPoster, error) {
	query := `
		SELECT
			fr.id, fr.film_id, fr.watched_date, fr.rating, fr.is_rewatch, fr.has_spoilers, fr.review_text, fr.published,
			f.id as film_id2, f.tmdb_id, f.title, f.year, f.overview, f.runtime, f.published as film_published
		FROM film_reviews fr
		JOIN films f ON fr.film_id = f.id
		WHERE fr.id = $1
	`

	var review FilmReview
	var film Film

	err := db.QueryRow(query, reviewID).Scan(
		&review.ID, &review.FilmID, &review.WatchedDate, &review.Rating, &review.IsRewatch, &review.HasSpoilers, &review.ReviewText, &review.Published,
		&film.ID, &film.TMDBID, &film.Title, &film.Year, &film.Overview, &film.Runtime, &film.Published,
	)
	if err != nil {
		return nil, err
	}

	result := FilmReviewWithFilmAndPoster{
		FilmReviewWithFilm: FilmReviewWithFilm{
			FilmReview: review,
			Film:       film,
		},
	}

	relations, err := GetMediaRelationsForEntity("film", film.ID)
	if err != nil {
		return nil, err
	}

	for _, r := range relations {
		if r.Role != nil && *r.Role == "poster" {
			result.Poster = &r
			break
		}
	}

	return &result, nil
}
