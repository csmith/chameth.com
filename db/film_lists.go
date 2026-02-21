package db

import (
	"context"
	"fmt"

	"chameth.com/chameth.com/metrics"
)

func GetAllFilmLists(ctx context.Context) ([]FilmList, error) {
	metrics.LogQuery(ctx)
	var lists []FilmList
	err := db.SelectContext(ctx, &lists, "SELECT id, title, description, published, path FROM film_lists WHERE published = true ORDER BY title")
	if err != nil {
		return nil, err
	}
	return lists, nil
}

func GetDraftFilmLists(ctx context.Context) ([]FilmList, error) {
	metrics.LogQuery(ctx)
	var lists []FilmList
	err := db.SelectContext(ctx, &lists, "SELECT id, title, description, published, path FROM film_lists WHERE published = false ORDER BY title")
	if err != nil {
		return nil, err
	}
	return lists, nil
}

func GetFilmListByID(ctx context.Context, id int) (*FilmList, error) {
	metrics.LogQuery(ctx)
	var list FilmList
	err := db.GetContext(ctx, &list, "SELECT id, title, description, published, path FROM film_lists WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

func GetFilmListByPath(ctx context.Context, path string) (*FilmList, error) {
	metrics.LogQuery(ctx)
	var list FilmList
	err := db.GetContext(ctx, &list, "SELECT id, title, description, published, path FROM film_lists WHERE path = $1 OR path = $2", path, path+"/")
	if err != nil {
		return nil, err
	}
	return &list, nil
}

func GetFilmListWithEntries(ctx context.Context, id int) (*FilmList, []FilmListEntryWithFilm, error) {
	metrics.LogQuery(ctx)
	list, err := GetFilmListByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	var entries []FilmListEntryWithFilm
	err = db.SelectContext(ctx, &entries, `
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
	metrics.LogQuery(ctx)
	var entries []FilmListEntryWithDetails
	err := db.SelectContext(ctx, &entries, `
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
	if err != nil {
		return nil, err
	}
	return entries, nil
}

func GetFilmListWithFilms(ctx context.Context, listID int) (*FilmList, int, []FilmListEntryWithPoster, error) {
	metrics.LogQuery(ctx)
	var list struct {
		ID          int    `db:"id"`
		Title       string `db:"title"`
		Description string `db:"description"`
		Published   bool   `db:"published"`
		Path        string `db:"path"`
		Count       int    `db:"count"`
	}
	err := db.GetContext(ctx, &list, `
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

	var entries []FilmListEntryWithPoster
	err = db.SelectContext(ctx, &entries, `
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

func CreateFilmList(ctx context.Context, path, title, description string) (int, error) {
	metrics.LogQuery(ctx)
	var id int
	err := db.QueryRowContext(ctx, `
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
	metrics.LogQuery(ctx)
	_, err := db.ExecContext(ctx, `
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
	metrics.LogQuery(ctx)
	var entries []FilmListEntryWithFilm
	err := db.SelectContext(ctx, &entries, `
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
	if err != nil {
		return nil, err
	}
	return entries, nil
}

func GetEntryByID(ctx context.Context, entryID int) (*FilmListEntry, error) {
	metrics.LogQuery(ctx)
	var entry FilmListEntry
	err := db.GetContext(ctx, &entry, "SELECT id, film_list_id, film_id, position FROM film_list_entries WHERE id = $1", entryID)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func AddFilmToList(ctx context.Context, listID, filmID int, position int) (int, error) {
	metrics.LogQuery(ctx)
	// Start a transaction to shift positions if needed
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

	// Shift all positions >= target position down by 1
	_, err = tx.Exec(`
		UPDATE film_list_entries
		SET position = position + 1
		WHERE film_list_id = $1 AND position >= $2
	`, listID, position)
	if err != nil {
		return 0, fmt.Errorf("failed to shift positions: %w", err)
	}

	// Now insert the new entry
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
	metrics.LogQuery(ctx)
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
	metrics.LogQuery(ctx)
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
	metrics.LogQuery(ctx)
	var position int
	err := db.GetContext(ctx, &position, `
		SELECT COALESCE(MAX(position), 0) + 1
		FROM film_list_entries
		WHERE film_list_id = $1
	`, listID)
	if err != nil {
		return 0, err
	}
	return position, nil
}

func GetFilmListsContainingFilm(ctx context.Context, filmID int) ([]FilmList, error) {
	metrics.LogQuery(ctx)
	var lists []FilmList
	err := db.SelectContext(ctx, &lists, `
		SELECT DISTINCT fl.id, fl.title, fl.description, fl.published, fl.path
		FROM film_lists fl
		JOIN film_list_entries fle ON fl.id = fle.film_list_id
		WHERE fl.published = true AND fle.film_id = $1
		ORDER BY fl.title
	`, filmID)
	if err != nil {
		return nil, err
	}
	return lists, nil
}

func ReorderFilmListEntries(ctx context.Context, listID int) error {
	metrics.LogQuery(ctx)
	_, err := db.ExecContext(ctx, `
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
