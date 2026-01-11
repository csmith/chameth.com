package db

import (
	"database/sql"
	"fmt"
	"strconv"
)

func GetFilmByID(id int) (*Film, error) {
	var film Film
	err := db.Get(&film, "SELECT id, tmdb_id, title, year, overview, runtime, published, path FROM films WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &film, nil
}

func GetAllFilms() ([]Film, error) {
	var films []Film
	err := db.Select(&films, "SELECT id, tmdb_id, title, year, overview, runtime, published, path FROM films ORDER BY title")
	if err != nil {
		return nil, err
	}
	return films, nil
}

func GetAllFilmsWithReviews() ([]FilmWithReview, error) {
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

	rows, err := db.Query(query)
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

func CreateFilm(tmdbID int, title, year, path string, overview string, runtime int) (int, error) {
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
	err := db.QueryRow(`
		INSERT INTO films (tmdb_id, title, year, overview, runtime, published, path)
		VALUES ($1, $2, $3, $4, $5, false, $6)
		RETURNING id
	`, tmdbID, title, yearPtr, overview, runtimePtr, path).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create film: %w", err)
	}
	return id, nil
}

func UpdateFilm(id int, tmdbID int, title, year, path string, overview string, runtime int, published bool) error {
	var tmdbIDPtr *int
	if tmdbID > 0 {
		tmdbIDPtr = &tmdbID
	}

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

	_, err := db.Exec(`
		UPDATE films
		SET tmdb_id = $1, title = $2, year = $3, overview = $4, runtime = $5, published = $6, path = $7
		WHERE id = $8
	`, tmdbIDPtr, title, yearPtr, overview, runtimePtr, published, path, id)
	if err != nil {
		return fmt.Errorf("failed to update film: %w", err)
	}
	return nil
}

func GetFilmByTMDBID(tmdbID int) (*Film, error) {
	var film Film
	err := db.Get(&film, "SELECT id, tmdb_id, title, year, overview, runtime, published, path FROM films WHERE tmdb_id = $1", tmdbID)
	if err != nil {
		return nil, err
	}
	return &film, nil
}

func GetFilmByPath(path string) (*Film, error) {
	var film Film
	err := db.Get(&film, "SELECT id, tmdb_id, title, year, overview, runtime, published, path FROM films WHERE path = $1 OR path = $2", path, path+"/")
	if err != nil {
		return nil, err
	}
	return &film, nil
}
