package films

import (
	"context"
	"fmt"
	"time"

	"chameth.com/chameth.com/db"
)

const filmColumns = `
	f.id, f.title, f.year, f.path, f.watch_count, f.average_rating, f.last_watched, mr.path AS poster_path
`

const posterJoin = `
	LEFT JOIN media_relations mr ON mr.entity_type = 'film' AND mr.entity_id = f.id AND mr.role = 'poster'
`

func getFilmByPath(ctx context.Context, path string) (Film, error) {
	return db.Get[Film](
		ctx,
		`SELECT `+filmColumns+` FROM films f `+posterJoin+` WHERE f.path = $1 OR f.path = $2`,
		path,
		path+"/",
	)
}

func searchFilms(ctx context.Context, query string) ([]Film, error) {
	return db.Select[Film](
		ctx,
		`SELECT `+filmColumns+` FROM films f `+posterJoin+`
		WHERE f.watch_count > 0 AND f.title ILIKE $1
		ORDER BY f.title, f.year`,
		"%"+query+"%",
	)
}

// SitemapFilms returns every film that has a published review.
func SitemapFilms(ctx context.Context) ([]Film, error) {
	return db.Select[Film](
		ctx,
		`SELECT `+filmColumns+` FROM films f `+posterJoin+`
		WHERE f.last_watched IS NOT NULL
		ORDER BY f.last_watched DESC, f.title`,
	)
}

const listColumns = `l.id, l.title, l.description, l.path, l.letterboxd_url, l.ranking`

func SitemapLists(ctx context.Context) ([]FilmList, error) {
	return db.Select[FilmList](ctx, `SELECT `+listColumns+` FROM film_lists l ORDER BY l.title`)
}

func getFilmListByPath(ctx context.Context, path string) (FilmList, error) {
	return db.Get[FilmList](
		ctx,
		`SELECT `+listColumns+` FROM film_lists l WHERE l.path = $1 OR l.path = $2`,
		path,
		path+"/",
	)
}

func getFilmList(ctx context.Context, id int) (FilmList, error) {
	return db.Get[FilmList](ctx, `SELECT `+listColumns+` FROM film_lists l WHERE l.id = $1`, id)
}

func countFilmListEntries(ctx context.Context, listID int) (int, error) {
	return db.Get[int](ctx, `SELECT COUNT(*) FROM film_list_entries WHERE film_list_id = $1`, listID)
}

// A limit of 0 returns every entry.
func getFilmListEntries(ctx context.Context, listID, limit int) ([]FilmListEntry, error) {
	var limitArg *int
	if limit > 0 {
		limitArg = &limit
	}
	return db.Select[FilmListEntry](
		ctx,
		`
		SELECT e.position,
			f.id AS "film.id", f.title AS "film.title", f.year AS "film.year", f.path AS "film.path",
			f.watch_count AS "film.watch_count", f.average_rating AS "film.average_rating",
			f.last_watched AS "film.last_watched", mr.path AS "film.poster_path"
		FROM film_list_entries e
		JOIN films f ON f.id = e.film_id
		`+posterJoin+`
		WHERE e.film_list_id = $1
		ORDER BY e.position
		LIMIT $2
	`,
		listID,
		limitArg,
	)
}

func getFilmListMemberships(ctx context.Context, filmID int) ([]FilmListMembership, error) {
	return db.Select[FilmListMembership](
		ctx,
		`
		SELECT `+listColumns+`, e.position,
			(SELECT COUNT(*) FROM film_list_entries t WHERE t.film_list_id = l.id) AS total
		FROM film_list_entries e
		JOIN film_lists l ON l.id = e.film_list_id
		WHERE e.film_id = $1
		ORDER BY l.title
	`,
		filmID,
	)
}

func getFilmSyncStates(ctx context.Context) ([]filmSyncState, error) {
	return db.Select[filmSyncState](ctx, `SELECT id, path, last_watched, poster_updated_at FROM films`)
}

func getListEntryRows(ctx context.Context) ([]listEntryRow, error) {
	return db.Select[listEntryRow](
		ctx,
		`SELECT film_list_id, film_id, position FROM film_list_entries ORDER BY film_list_id, position`,
	)
}

// replaceIndex makes the mirrored films, lists and (when entries is non-nil)
// list entries match the given index, in a single transaction.
func replaceIndex(ctx context.Context, idx *index, entries []listEntryRow) (err error) {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	filmIDs := make([]int, len(idx.Films))
	for i, f := range idx.Films {
		filmIDs[i] = f.ID
	}
	listIDs := make([]int, len(idx.Lists))
	for i, l := range idx.Lists {
		listIDs[i] = l.ID
	}

	// Removals go first so a path freed by one row can be taken by another.
	if _, err = tx.ExecContext(ctx, `DELETE FROM films WHERE NOT (id = ANY($1))`, filmIDs); err != nil {
		return fmt.Errorf("failed to delete films: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM film_lists WHERE NOT (id = ANY($1))`, listIDs); err != nil {
		return fmt.Errorf("failed to delete film lists: %w", err)
	}

	for _, f := range idx.Films {
		_, err = tx.ExecContext(
			ctx,
			`
			INSERT INTO films (id, title, year, path, watch_count, average_rating, last_watched)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id) DO UPDATE SET
				title = EXCLUDED.title,
				year = EXCLUDED.year,
				path = EXCLUDED.path,
				watch_count = EXCLUDED.watch_count,
				average_rating = EXCLUDED.average_rating,
				last_watched = EXCLUDED.last_watched
			WHERE (films.title, films.year, films.path, films.watch_count, films.average_rating, films.last_watched)
				IS DISTINCT FROM
				(EXCLUDED.title, EXCLUDED.year, EXCLUDED.path, EXCLUDED.watch_count, EXCLUDED.average_rating, EXCLUDED.last_watched)
		`,
			f.ID,
			f.Title,
			f.Year,
			FilmPath(f.Slug),
			f.WatchCount,
			f.AverageRating,
			f.LastWatched,
		)
		if err != nil {
			return fmt.Errorf("failed to upsert film %d: %w", f.ID, err)
		}
	}

	for _, l := range idx.Lists {
		_, err = tx.ExecContext(
			ctx,
			`
			INSERT INTO film_lists (id, title, description, path, letterboxd_url, ranking)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (id) DO UPDATE SET
				title = EXCLUDED.title,
				description = EXCLUDED.description,
				path = EXCLUDED.path,
				letterboxd_url = EXCLUDED.letterboxd_url,
				ranking = EXCLUDED.ranking
			WHERE (film_lists.title, film_lists.description, film_lists.path, film_lists.letterboxd_url, film_lists.ranking)
				IS DISTINCT FROM
				(EXCLUDED.title, EXCLUDED.description, EXCLUDED.path, EXCLUDED.letterboxd_url, EXCLUDED.ranking)
		`,
			l.ID,
			l.Title,
			l.Description,
			listPath(l.Slug),
			l.LetterboxdURL,
			l.Ranking,
		)
		if err != nil {
			return fmt.Errorf("failed to upsert film list %d: %w", l.ID, err)
		}
	}

	if entries != nil {
		if _, err = tx.ExecContext(ctx, `DELETE FROM film_list_entries`); err != nil {
			return fmt.Errorf("failed to clear film list entries: %w", err)
		}
		for _, e := range entries {
			_, err = tx.ExecContext(
				ctx,
				`INSERT INTO film_list_entries (film_list_id, film_id, position) VALUES ($1, $2, $3)`,
				e.ListID,
				e.FilmID,
				e.Position,
			)
			if err != nil {
				return fmt.Errorf("failed to insert film list entry: %w", err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func setPosterUpdatedAt(ctx context.Context, filmID int, updatedAt time.Time) error {
	_, err := db.Exec(ctx, `UPDATE films SET poster_updated_at = $1 WHERE id = $2`, updatedAt, filmID)
	return err
}

func getPosterRelations(ctx context.Context, filmID int) ([]posterRelation, error) {
	return db.Select[posterRelation](
		ctx,
		`
		SELECT path, media_id, entity_id
		FROM media_relations
		WHERE entity_type = 'film' AND role = 'poster' AND entity_id = $1
	`,
		filmID,
	)
}

func getOrphanedPosterRelations(ctx context.Context) ([]posterRelation, error) {
	return db.Select[posterRelation](
		ctx,
		`
		SELECT path, media_id, entity_id
		FROM media_relations
		WHERE entity_type = 'film' AND entity_id NOT IN (SELECT id FROM films)
	`,
	)
}
