package films

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"chameth.com/chameth.com/db"
)

func GetFilmByID(ctx context.Context, id int) (*Film, error) {
	film, err := db.Get[Film](ctx, "SELECT id, tmdb_id, title, year, overview, runtime, published, path FROM films WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &film, nil
}

func GetAllFilms(ctx context.Context) ([]Film, error) {
	return db.Select[Film](ctx, "SELECT id, tmdb_id, title, year, overview, runtime, published, path FROM films ORDER BY title")
}

func GetAllFilmsWithReviews(ctx context.Context) ([]FilmWithReview, error) {
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

	rows, err := db.Query(ctx, query)
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

	rows, err := db.Query(ctx, query)
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
	err := db.QueryRow(ctx, `
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

	_, err := db.Exec(ctx, `
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
	film, err := db.Get[Film](ctx, "SELECT id, tmdb_id, title, year, overview, runtime, published, path FROM films WHERE tmdb_id = $1", tmdbID)
	if err != nil {
		return nil, err
	}
	return &film, nil
}

func GetFilmWithPosterByPath(ctx context.Context, path string) (*FilmWithPoster, error) {
	film, err := db.Get[FilmWithPoster](ctx, `
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
	_, err := db.Exec(ctx, "DELETE FROM films WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete film: %w", err)
	}
	return nil
}

func SearchFilms(ctx context.Context, query string) ([]FilmSearchResult, error) {
	return db.Select[FilmSearchResult](ctx, `
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
}

func GetFilmReviewByID(ctx context.Context, id int) (*FilmReview, error) {
	review, err := db.Get[FilmReview](ctx, "SELECT id, film_id, watched_date, rating, is_rewatch, has_spoilers, review_text, published FROM film_reviews WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func GetFilmReviewsByFilmID(ctx context.Context, filmID int) ([]FilmReview, error) {
	return db.Select[FilmReview](ctx, "SELECT id, film_id, watched_date, rating, is_rewatch, has_spoilers, review_text, published FROM film_reviews WHERE film_id = $1 ORDER BY watched_date DESC", filmID)
}

func CreateFilmReview(ctx context.Context, filmID int, rating int, watchedDate any, isRewatch, hasSpoilers, published bool, reviewText string) (int, error) {
	var id int
	err := db.QueryRow(ctx, `
		INSERT INTO film_reviews (film_id, rating, watched_date, is_rewatch, has_spoilers, review_text, published)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, filmID, rating, watchedDate, isRewatch, hasSpoilers, reviewText, published).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create film review: %w", err)
	}
	return id, nil
}

func DeleteFilmReview(ctx context.Context, id int) error {
	_, err := db.Exec(ctx, "DELETE FROM film_reviews WHERE id = $1 AND published = false", id)
	if err != nil {
		return fmt.Errorf("failed to delete film review: %w", err)
	}
	return nil
}

func UpdateFilmReview(ctx context.Context, id int, rating int, watchedDate string, isRewatch, hasSpoilers, published bool, reviewText string) error {
	_, err := db.Exec(ctx, `
		UPDATE film_reviews
		SET rating = $1, watched_date = $2, is_rewatch = $3, has_spoilers = $4, review_text = $5, published = $6
		WHERE id = $7
	`, rating, watchedDate, isRewatch, hasSpoilers, reviewText, published, id)
	if err != nil {
		return fmt.Errorf("failed to update film review: %w", err)
	}
	return nil
}

func GetAllPublishedFilmReviewsWithFilmAndPosters(ctx context.Context) ([]FilmReviewWithFilmAndPoster, error) {
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

	return db.Select[FilmReviewWithFilmAndPoster](ctx, query)
}

func GetRecentPublishedFilmReviewsWithFilmAndPosters(ctx context.Context, limit int) ([]FilmReviewWithFilmAndPoster, error) {
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

	return db.Select[FilmReviewWithFilmAndPoster](ctx, query, limit)
}

func GetAllFilmLists(ctx context.Context) ([]FilmList, error) {
	return db.Select[FilmList](ctx, "SELECT id, title, description, published, path FROM film_lists WHERE published = true ORDER BY title")
}

func GetDraftFilmLists(ctx context.Context) ([]FilmList, error) {
	return db.Select[FilmList](ctx, "SELECT id, title, description, published, path FROM film_lists WHERE published = false ORDER BY title")
}

func GetFilmListByID(ctx context.Context, id int) (*FilmList, error) {
	list, err := db.Get[FilmList](ctx, "SELECT id, title, description, published, path FROM film_lists WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

func GetFilmListByPath(ctx context.Context, path string) (*FilmList, error) {
	list, err := db.Get[FilmList](ctx, "SELECT id, title, description, published, path FROM film_lists WHERE path = $1 OR path = $2", path, path+"/")
	if err != nil {
		return nil, err
	}
	return &list, nil
}

func GetFilmListWithEntries(ctx context.Context, id int) (*FilmList, []FilmListEntryWithFilm, error) {
	list, err := GetFilmListByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	entries, err := db.Select[FilmListEntryWithFilm](ctx, `
		SELECT
			fle.id, fle.film_list_id, fle.film_id, fle.position,
			f.id as "film.id", f.tmdb_id as "film.tmdb_id", f.title as "film.title",
			f.year as "film.year", f.overview as "film.overview", f.runtime as "film.runtime",
			f.published as "film.published", f.path as "film.path"
		FROM film_list_entries fle
		JOIN films f ON fle.film_id = f.id
		WHERE fle.film_list_id = $1
		ORDER BY fle.position
	`, id)
	if err != nil {
		return nil, nil, err
	}

	return list, entries, nil
}

func GetFilmListEntriesWithDetails(ctx context.Context, listID int) ([]FilmListEntryWithDetails, error) {
	return db.Select[FilmListEntryWithDetails](ctx, `
		SELECT
			fle.id, fle.film_list_id, fle.film_id, fle.position,
			f.id as "film.id", f.tmdb_id as "film.tmdb_id", f.title as "film.title",
			f.year as "film.year", f.overview as "film.overview", f.runtime as "film.runtime",
			f.published as "film.published", f.path as "film.path",
			COUNT(fr.id) as times_watched,
			AVG(fr.rating) as average_rating,
			MAX(fr.watched_date) as last_watched,
			mr.path as "poster.path", mr.media_id as "poster.media_id", mr.description as "poster.description",
			mr.caption as "poster.caption", mr.role as "poster.role", mr.entity_type as "poster.entity_type",
			mr.entity_id as "poster.entity_id"
		FROM film_list_entries fle
		JOIN films f ON fle.film_id = f.id
		LEFT JOIN film_reviews fr ON fr.film_id = f.id AND fr.published = true
		JOIN media_relations mr ON mr.entity_type = 'film' AND mr.entity_id = f.id AND mr.role = 'poster'
		WHERE fle.film_list_id = $1
		GROUP BY fle.id, f.id, mr.path, mr.media_id, mr.description, mr.caption, mr.role, mr.entity_type, mr.entity_id
		ORDER BY fle.position
	`, listID)
}

func CreateFilmList(ctx context.Context, path, title, description string) (int, error) {
	var id int
	err := db.QueryRow(ctx, `
		INSERT INTO film_lists (path, title, description, published)
		VALUES ($1, $2, $3, false)
		RETURNING id
	`, path, title, description).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create film list: %w", err)
	}
	return id, nil
}

func UpdateFilmList(ctx context.Context, id int, path, title, description string, published bool) error {
	_, err := db.Exec(ctx, `
		UPDATE film_lists
		SET path = $1, title = $2, description = $3, published = $4
		WHERE id = $5
	`, path, title, description, published, id)
	if err != nil {
		return fmt.Errorf("failed to update film list: %w", err)
	}
	return nil
}

func GetEntriesForList(ctx context.Context, listID int) ([]FilmListEntryWithFilm, error) {
	return db.Select[FilmListEntryWithFilm](ctx, `
		SELECT
			fle.id, fle.film_list_id, fle.film_id, fle.position,
			f.id as "film.id", f.tmdb_id as "film.tmdb_id", f.title as "film.title",
			f.year as "film.year", f.overview as "film.overview", f.runtime as "film.runtime",
			f.published as "film.published", f.path as "film.path"
		FROM film_list_entries fle
		JOIN films f ON fle.film_id = f.id
		WHERE fle.film_list_id = $1
		ORDER BY fle.position
	`, listID)
}

func GetEntryByID(ctx context.Context, entryID int) (*FilmListEntry, error) {
	entry, err := db.Get[FilmListEntry](ctx, "SELECT id, film_list_id, film_id, position FROM film_list_entries WHERE id = $1", entryID)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func AddFilmToList(ctx context.Context, listID, filmID int, position int) (int, error) {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, err = tx.Exec("SET CONSTRAINTS ALL DEFERRED")
	if err != nil {
		return 0, fmt.Errorf("failed to defer constraints: %w", err)
	}

	_, err = tx.Exec(`
		UPDATE film_list_entries
		SET position = position + 1
		WHERE film_list_id = $1 AND position >= $2
	`, listID, position)
	if err != nil {
		return 0, fmt.Errorf("failed to shift positions: %w", err)
	}

	var id int
	err = tx.QueryRow(`
		INSERT INTO film_list_entries (film_list_id, film_id, position)
		VALUES ($1, $2, $3)
		RETURNING id
	`, listID, filmID, position).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to add film to list: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return id, nil
}

func RemoveFilmFromList(ctx context.Context, entryID int) error {
	entry, err := GetEntryByID(ctx, entryID)
	if err != nil {
		return fmt.Errorf("failed to get entry: %w", err)
	}

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, err = tx.Exec("SET CONSTRAINTS ALL DEFERRED")
	if err != nil {
		return fmt.Errorf("failed to defer constraints: %w", err)
	}

	_, err = tx.Exec(`
		UPDATE film_list_entries
		SET position = position - 1
		WHERE film_list_id = $1 AND position > $2
	`, entry.FilmListID, entry.Position)
	if err != nil {
		return fmt.Errorf("failed to reflow positions: %w", err)
	}

	_, err = tx.Exec("DELETE FROM film_list_entries WHERE id = $1", entryID)
	if err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func UpdateEntryPosition(ctx context.Context, entryID, newPosition int) error {
	entry, err := GetEntryByID(ctx, entryID)
	if err != nil {
		return fmt.Errorf("failed to get entry: %w", err)
	}

	oldPosition := entry.Position

	if oldPosition == newPosition {
		return nil
	}

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, err = tx.Exec("SET CONSTRAINTS ALL DEFERRED")
	if err != nil {
		return fmt.Errorf("failed to defer constraints: %w", err)
	}

	if oldPosition < newPosition {
		_, err = tx.Exec(`
			UPDATE film_list_entries
			SET position = position - 1
			WHERE film_list_id = $1 AND position > $2 AND position <= $3
		`, entry.FilmListID, oldPosition, newPosition)
		if err != nil {
			return fmt.Errorf("failed to shift positions down: %w", err)
		}
	} else {
		_, err = tx.Exec(`
			UPDATE film_list_entries
			SET position = position + 1
			WHERE film_list_id = $1 AND position >= $2 AND position < $3
		`, entry.FilmListID, newPosition, oldPosition)
		if err != nil {
			return fmt.Errorf("failed to shift positions up: %w", err)
		}
	}

	_, err = tx.Exec(`
		UPDATE film_list_entries
		SET position = $1
		WHERE id = $2
	`, newPosition, entryID)
	if err != nil {
		return fmt.Errorf("failed to update entry position: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func GetNextPosition(ctx context.Context, listID int) (int, error) {
	return db.Get[int](ctx, `
		SELECT COALESCE(MAX(position), 0) + 1
		FROM film_list_entries
		WHERE film_list_id = $1
	`, listID)
}

func GetFilmListsContainingFilm(ctx context.Context, filmID int) ([]FilmList, error) {
	return db.Select[FilmList](ctx, `
		SELECT DISTINCT fl.id, fl.title, fl.description, fl.published, fl.path
		FROM film_lists fl
		JOIN film_list_entries fle ON fl.id = fle.film_list_id
		WHERE fl.published = true AND fle.film_id = $1
		ORDER BY fl.title
	`, filmID)
}

func ReorderFilmListEntries(ctx context.Context, listID int) error {
	_, err := db.Exec(ctx, `
		UPDATE film_list_entries AS fle
		SET position = sub.rn
		FROM (
			SELECT id, row_number() OVER (ORDER BY position) as rn
			FROM film_list_entries
			WHERE film_list_id = $1
		) AS sub
		WHERE fle.id = sub.id
	`, listID)
	if err != nil {
		return fmt.Errorf("failed to reorder film list entries: %w", err)
	}
	return nil
}

func GetRatingDistribution(ctx context.Context) ([]FilmRatingDistribution, error) {
	return db.Select[FilmRatingDistribution](ctx, `
		SELECT rating, COUNT(*) as count
		FROM film_reviews
		WHERE published = true
		GROUP BY rating
		ORDER BY rating ASC
	`)
}

func GetFilmReviewWithFilmAndPoster(ctx context.Context, reviewID int) (*FilmReviewWithFilmAndPoster, error) {
	result, err := db.Get[FilmReviewWithFilmAndPoster](ctx, `
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
	`, reviewID)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func GetFilmListWithCount(ctx context.Context, listID int) (*FilmList, int, []FilmListEntryWithPoster, error) {
	list, err := db.Get[struct {
		ID          int    `db:"id"`
		Title       string `db:"title"`
		Description string `db:"description"`
		Published   bool   `db:"published"`
		Path        string `db:"path"`
		Count       int    `db:"count"`
	}](ctx, `
		SELECT
			fl.id, fl.title, fl.description, fl.published, fl.path,
			COUNT(fle.id) as count
		FROM film_lists fl
		LEFT JOIN film_list_entries fle ON fl.id = fle.film_list_id
		WHERE fl.id = $1
		GROUP BY fl.id
	`, listID)
	if err != nil {
		return nil, 0, nil, err
	}

	entries, err := db.Select[FilmListEntryWithPoster](ctx, `
		SELECT
			fle.id, fle.position,
			f.id as "film.id", f.tmdb_id as "film.tmdb_id", f.title as "film.title",
			f.year as "film.year", f.overview as "film.overview", f.runtime as "film.runtime",
			f.published as "film.published", f.path as "film.path",
			mr.path as "poster.path", mr.media_id as "poster.media_id"
		FROM film_list_entries fle
		JOIN films f ON fle.film_id = f.id
		LEFT JOIN media_relations mr ON mr.entity_type = 'film' AND mr.entity_id = f.id AND mr.role = 'poster'
		WHERE fle.film_list_id = $1
		ORDER BY fle.position
		LIMIT 5
	`, listID)
	if err != nil {
		return nil, 0, nil, err
	}

	filmList := &FilmList{
		ID:          list.ID,
		Title:       list.Title,
		Description: list.Description,
		Published:   list.Published,
		Path:        list.Path,
	}

	return filmList, list.Count, entries, nil
}

func GetWatchedFilmsByDateRange(ctx context.Context, startDate, endDate any) ([]FilmReviewWithFilmAndPoster, error) {
	return db.Select[FilmReviewWithFilmAndPoster](ctx, `
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
