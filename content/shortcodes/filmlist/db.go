package filmlist

import (
	"context"

	"chameth.com/chameth.com/db"
)

func query(ctx context.Context, listID int) (*db.FilmList, int, []db.FilmListEntryWithPoster, error) {
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

	entries, err := db.Select[db.FilmListEntryWithPoster](ctx, `
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

	filmList := &db.FilmList{
		ID:          list.ID,
		Title:       list.Title,
		Description: list.Description,
		Published:   list.Published,
		Path:        list.Path,
	}

	return filmList, list.Count, entries, nil
}
