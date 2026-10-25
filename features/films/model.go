package films

import (
	"time"
)

type Film struct {
	ID            int        `db:"id"`
	Title         string     `db:"title"`
	Year          *int       `db:"year"`
	Path          string     `db:"path"`
	WatchCount    int        `db:"watch_count"`
	AverageRating *float64   `db:"average_rating"`
	LastWatched   *time.Time `db:"last_watched"`
	PosterPath    *string    `db:"poster_path"`
}

type FilmList struct {
	ID            int    `db:"id"`
	Title         string `db:"title"`
	Description   string `db:"description"`
	Path          string `db:"path"`
	LetterboxdURL string `db:"letterboxd_url"`
	Ranking       bool   `db:"ranking"`
}

type FilmListEntry struct {
	Position int  `db:"position"`
	Film     Film `db:"film"`
}

type FilmListMembership struct {
	FilmList
	Position int `db:"position"`
	Total    int `db:"total"`
}

type FilmListCard struct {
	FilmList
	Count int
	Films []Film
}

type filmSyncState struct {
	ID              int        `db:"id"`
	Path            string     `db:"path"`
	LastWatched     *time.Time `db:"last_watched"`
	PosterUpdatedAt *time.Time `db:"poster_updated_at"`
}

type listEntryRow struct {
	ListID   int `db:"film_list_id"`
	FilmID   int `db:"film_id"`
	Position int `db:"position"`
}

type filmMedia struct {
	Path     string `db:"path"`
	MediaID  int    `db:"media_id"`
	EntityID int    `db:"entity_id"`
}
