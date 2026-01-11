package db

import (
	"fmt"
)

func GetAllFilmLists() ([]FilmList, error) {
	var lists []FilmList
	err := db.Select(&lists, "SELECT id, title, description, published, path FROM film_lists WHERE published = true ORDER BY title")
	if err != nil {
		return nil, err
	}
	return lists, nil
}

func GetDraftFilmLists() ([]FilmList, error) {
	var lists []FilmList
	err := db.Select(&lists, "SELECT id, title, description, published, path FROM film_lists WHERE published = false ORDER BY title")
	if err != nil {
		return nil, err
	}
	return lists, nil
}

func GetFilmListByID(id int) (*FilmList, error) {
	var list FilmList
	err := db.Get(&list, "SELECT id, title, description, published, path FROM film_lists WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

func GetFilmListByPath(path string) (*FilmList, error) {
	var list FilmList
	err := db.Get(&list, "SELECT id, title, description, published, path FROM film_lists WHERE path = $1 OR path = $2", path, path+"/")
	if err != nil {
		return nil, err
	}
	return &list, nil
}

func GetFilmListWithEntries(id int) (*FilmList, []FilmListEntryWithFilm, error) {
	list, err := GetFilmListByID(id)
	if err != nil {
		return nil, nil, err
	}

	var entries []FilmListEntryWithFilm
	err = db.Select(&entries, `
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

func GetFilmListEntriesWithDetails(listID int) ([]FilmListEntryWithDetails, error) {
	var entries []FilmListEntryWithDetails
	err := db.Select(&entries, `
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

func GetFilmListWithFilms(listID int) (*FilmList, int, []FilmListEntryWithPoster, error) {
	var list struct {
		ID          int    `db:"id"`
		Title       string `db:"title"`
		Description string `db:"description"`
		Published   bool   `db:"published"`
		Path        string `db:"path"`
		Count       int    `db:"count"`
	}
	err := db.Get(&list, `
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
	err = db.Select(&entries, `
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

func CreateFilmList(path, title, description string) (int, error) {
	var id int
	err := db.QueryRow(`
		INSERT INTO film_lists (path, title, description, published)
		VALUES ($1, $2, $3, false)
		RETURNING id
	`, path, title, description).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create film list: %w", err)
	}
	return id, nil
}

func UpdateFilmList(id int, path, title, description string, published bool) error {
	_, err := db.Exec(`
		UPDATE film_lists
		SET path = $1, title = $2, description = $3, published = $4
		WHERE id = $5
	`, path, title, description, published, id)
	if err != nil {
		return fmt.Errorf("failed to update film list: %w", err)
	}
	return nil
}

func GetEntriesForList(listID int) ([]FilmListEntryWithFilm, error) {
	var entries []FilmListEntryWithFilm
	err := db.Select(&entries, `
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

func GetEntryByID(entryID int) (*FilmListEntry, error) {
	var entry FilmListEntry
	err := db.Get(&entry, "SELECT id, film_list_id, film_id, position FROM film_list_entries WHERE id = $1", entryID)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func AddFilmToList(listID, filmID int, position int) (int, error) {
	var id int
	err := db.QueryRow(`
		INSERT INTO film_list_entries (film_list_id, film_id, position)
		VALUES ($1, $2, $3)
		RETURNING id
	`, listID, filmID, position).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to add film to list: %w", err)
	}
	return id, nil
}

func RemoveFilmFromList(entryID int) error {
	entry, err := GetEntryByID(entryID)
	if err != nil {
		return fmt.Errorf("failed to get entry: %w", err)
	}

	tx, err := db.Begin()
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

func UpdateEntryPosition(entryID, newPosition int) error {
	entry, err := GetEntryByID(entryID)
	if err != nil {
		return fmt.Errorf("failed to get entry: %w", err)
	}

	oldPosition := entry.Position

	if oldPosition == newPosition {
		return nil
	}

	tx, err := db.Begin()
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

func GetNextPosition(listID int) (int, error) {
	var position int
	err := db.Get(&position, `
		SELECT COALESCE(MAX(position), 0) + 1
		FROM film_list_entries
		WHERE film_list_id = $1
	`, listID)
	if err != nil {
		return 0, err
	}
	return position, nil
}
