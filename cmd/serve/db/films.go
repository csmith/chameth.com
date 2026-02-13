package db

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"chameth.com/chameth.com/cmd/serve/metrics"
)

func GetFilmByID(ctx context.Context, id int) (*Film, error) {
	metrics.LogQuery(ctx)
	var film Film
	err := db.GetContext(ctx, &film, "SELECT id, tmdb_id, title, year, overview, runtime, published, path FROM films WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &film, nil
}

func GetAllFilms(ctx context.Context) ([]Film, error) {
	metrics.LogQuery(ctx)
	var films []Film
	err := db.SelectContext(ctx, &films, "SELECT id, tmdb_id, title, year, overview, runtime, published, path FROM films ORDER BY title")
	if err != nil {
		return nil, err
	}
	return films, nil
}

func GetAllFilmsWithReviews(ctx context.Context) ([]FilmWithReview, error) {
	metrics.LogQuery(ctx)
	query := `
		SELECT
			f.id, f.tmdb_id, f.title, f.year, f.overview, f.runtime, f.published, f.path,
			fr.id as review_id, fr.film_id as review_film_id, fr.watched_date, fr.rating, fr.is_rewatch, fr.has_spoilers, fr.review_text, fr.published as review_published
		FROM films f
		LEFT JOIN LATERAL (
			SELECT * FROM film_reviews
			WHERE film_id = f.id
			ORDER BY watched_date DESC
			LIMIT 1
		) fr ON true
		ORDER BY f.title
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var films []FilmWithReview
	for rows.Next() {
		var f Film
		var review FilmReview
		var reviewID, reviewFilmID sql.NullInt64
		var watchedDate sql.NullTime
		var rating sql.NullInt64
		var reviewText sql.NullString
		var reviewPublished sql.NullBool

		err := rows.Scan(
			&f.ID, &f.TMDBID, &f.Title, &f.Year, &f.Overview, &f.Runtime, &f.Published, &f.Path,
			&reviewID, &reviewFilmID, &watchedDate, &rating, &review.IsRewatch, &review.HasSpoilers, &reviewText, &reviewPublished,
		)
		if err != nil {
			return nil, err
		}

		if reviewID.Valid {
			review.ID = int(reviewID.Int64)
			review.FilmID = int(reviewFilmID.Int64)
			review.WatchedDate = watchedDate.Time
			review.Rating = int(rating.Int64)
			review.ReviewText = reviewText.String
			review.Published = reviewPublished.Bool
			films = append(films, FilmWithReview{Film: f, Review: &review})
		} else {
			films = append(films, FilmWithReview{Film: f, Review: nil})
		}
	}

	return films, nil
}

func GetAllFilmsWithReviewsAndPosters(ctx context.Context) ([]FilmWithReviewAndPoster, error) {
	metrics.LogQuery(ctx)
	query := `
		SELECT
			f.id, f.tmdb_id, f.title, f.year, f.overview, f.runtime, f.published, f.path,
			fr.id as review_id, fr.film_id as review_film_id, fr.watched_date, fr.rating, fr.is_rewatch, fr.has_spoilers, fr.review_text, fr.published as review_published,
			mr.path as poster_path, mr.media_id as poster_media_id,
			(SELECT COUNT(*) FROM film_reviews WHERE film_id = f.id) as review_count,
			(SELECT to_char(MAX(watched_date), 'YYYY-MM-DD') FROM film_reviews WHERE film_id = f.id) as last_watched
		FROM films f
		LEFT JOIN LATERAL (
			SELECT * FROM film_reviews
			WHERE film_id = f.id
			ORDER BY watched_date DESC
			LIMIT 1
		) fr ON true
		LEFT JOIN media_relations mr ON mr.entity_type = 'film' AND mr.entity_id = f.id AND mr.role = 'poster'
		ORDER BY f.title
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var films []FilmWithReviewAndPoster
	for rows.Next() {
		var f Film
		var review FilmReview
		var reviewID, reviewFilmID sql.NullInt64
		var watchedDate sql.NullTime
		var rating sql.NullInt64
		var reviewText sql.NullString
		var reviewPublished sql.NullBool
		var posterPath sql.NullString
		var posterMediaID sql.NullInt64
		var reviewCount int
		var lastWatched sql.NullString

		err := rows.Scan(
			&f.ID, &f.TMDBID, &f.Title, &f.Year, &f.Overview, &f.Runtime, &f.Published, &f.Path,
			&reviewID, &reviewFilmID, &watchedDate, &rating, &review.IsRewatch, &review.HasSpoilers, &reviewText, &reviewPublished,
			&posterPath, &posterMediaID,
			&reviewCount, &lastWatched,
		)
		if err != nil {
			return nil, err
		}

		fwr := FilmWithReview{Film: f}
		if reviewID.Valid {
			review.ID = int(reviewID.Int64)
			review.FilmID = int(reviewFilmID.Int64)
			review.WatchedDate = watchedDate.Time
			review.Rating = int(rating.Int64)
			review.ReviewText = reviewText.String
			review.Published = reviewPublished.Bool
			fwr.Review = &review
		}

		var pp *string
		var pmi *int
		if posterPath.Valid {
			pp = &posterPath.String
			mid := int(posterMediaID.Int64)
			pmi = &mid
		}

		var lw *string
		if lastWatched.Valid {
			lw = &lastWatched.String
		}

		films = append(films, FilmWithReviewAndPoster{FilmWithReview: fwr, PosterPath: pp, PosterMediaID: pmi, ReviewCount: reviewCount, LastWatched: lw})
	}

	return films, nil
}

func CreateFilm(ctx context.Context, tmdbID int, title, year, path string, overview string, runtime int) (int, error) {
	metrics.LogQuery(ctx)
	var yearPtr *int
	if year != "" {
		y, err := strconv.Atoi(year)
		if err == nil {
			yearPtr = &y
		}
	}

	var runtimePtr *int
	if runtime > 0 {
		runtimePtr = &runtime
	}

	var id int
	err := db.QueryRowContext(ctx, `
		INSERT INTO films (tmdb_id, title, year, overview, runtime, published, path)
		VALUES ($1, $2, $3, $4, $5, false, $6)
		RETURNING id
	`, tmdbID, title, yearPtr, overview, runtimePtr, path).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create film: %w", err)
	}
	return id, nil
}

func UpdateFilm(ctx context.Context, id int, tmdbID *int, title, year, path string, overview string, runtime int, published bool) error {
	metrics.LogQuery(ctx)
	var yearPtr *int
	if year != "" {
		y, err := strconv.Atoi(year)
		if err == nil {
			yearPtr = &y
		}
	}

	var runtimePtr *int
	if runtime > 0 {
		runtimePtr = &runtime
	}

	_, err := db.ExecContext(ctx, `
		UPDATE films
		SET tmdb_id = $1, title = $2, year = $3, overview = $4, runtime = $5, published = $6, path = $7
		WHERE id = $8
	`, tmdbID, title, yearPtr, overview, runtimePtr, published, path, id)
	if err != nil {
		return fmt.Errorf("failed to update film: %w", err)
	}
	return nil
}

func GetFilmByTMDBID(ctx context.Context, tmdbID int) (*Film, error) {
	metrics.LogQuery(ctx)
	var film Film
	err := db.GetContext(ctx, &film, "SELECT id, tmdb_id, title, year, overview, runtime, published, path FROM films WHERE tmdb_id = $1", tmdbID)
	if err != nil {
		return nil, err
	}
	return &film, nil
}

func GetFilmWithPosterByPath(ctx context.Context, path string) (*FilmWithPoster, error) {
	metrics.LogQuery(ctx)
	var film FilmWithPoster
	err := db.GetContext(ctx, &film, `
		SELECT
			f.id, f.tmdb_id, f.title, f.year, f.overview, f.runtime, f.published, f.path,
			mr.path as poster_path
		FROM films f
		LEFT JOIN media_relations mr ON mr.entity_type = 'film' AND mr.entity_id = f.id AND mr.role = 'poster'
		WHERE f.path = $1 OR f.path = $2
	`, path, path+"/")
	if err != nil {
		return nil, err
	}
	return &film, nil
}

func DeleteFilm(ctx context.Context, id int) error {
	metrics.LogQuery(ctx)
	_, err := db.ExecContext(ctx, "DELETE FROM films WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete film: %w", err)
	}
	return nil
}

func SearchFilms(ctx context.Context, query string) ([]FilmSearchResult, error) {
	metrics.LogQuery(ctx)
	var results []FilmSearchResult
	err := db.SelectContext(ctx, &results, `
		SELECT
			f.id,
			f.title,
			f.path,
			mr.path as poster_path,
			(SELECT COUNT(*) FROM film_reviews WHERE film_id = f.id AND published = true) as times_watched,
			(SELECT AVG(rating) FROM film_reviews WHERE film_id = f.id AND published = true) as average_rating
		FROM films f
		LEFT JOIN media_relations mr ON mr.entity_type = 'film' AND mr.entity_id = f.id AND mr.role = 'poster'
		WHERE EXISTS (SELECT 1 FROM film_reviews WHERE film_id = f.id AND published = true)
			AND f.title ILIKE $1
		ORDER BY f.title
	`, "%"+query+"%")
	if err != nil {
		return nil, err
	}
	return results, nil
}
