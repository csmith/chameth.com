package films

import (
	"time"

	"chameth.com/chameth.com/features/media"
)

type FilmMetadata struct {
	ID        int    `db:"id"`
	TMDBID    *int   `db:"tmdb_id"`
	Title     string `db:"title"`
	Year      *int   `db:"year"`
	Published bool   `db:"published"`
	Path      string `db:"path"`
}

type Film struct {
	FilmMetadata
	Overview string `db:"overview"`
	Runtime  *int   `db:"runtime"`
}

type FilmWithPoster struct {
	ID         int     `db:"id"`
	TMDBID     *int    `db:"tmdb_id"`
	Title      string  `db:"title"`
	Year       *int    `db:"year"`
	Overview   string  `db:"overview"`
	Runtime    *int    `db:"runtime"`
	Published  bool    `db:"published"`
	Path       string  `db:"path"`
	PosterPath *string `db:"poster_path"`
}

type FilmReview struct {
	ID          int       `db:"id"`
	FilmID      int       `db:"film_id"`
	WatchedDate time.Time `db:"watched_date"`
	Rating      int       `db:"rating"`
	IsRewatch   bool      `db:"is_rewatch"`
	HasSpoilers bool      `db:"has_spoilers"`
	ReviewText  string    `db:"review_text"`
	Published   bool      `db:"published"`
}

type FilmWithReview struct {
	Film
	Review *FilmReview
}

type FilmWithReviewAndPoster struct {
	FilmWithReview
	PosterPath    *string
	PosterMediaID *int
	ReviewCount   int
	LastWatched   *string
}

type FilmReviewWithFilmAndPoster struct {
	FilmReview `db:"filmreview"`
	Film       `db:"film"`
	Poster     media.MediaRelationWithDetails `db:"poster"`
}

type FilmList struct {
	ID          int    `db:"id"`
	Title       string `db:"title"`
	Description string `db:"description"`
	Published   bool   `db:"published"`
	Path        string `db:"path"`
}

type FilmListEntry struct {
	ID         int `db:"id"`
	FilmListID int `db:"film_list_id"`
	FilmID     int `db:"film_id"`
	Position   int `db:"position"`
}

type FilmListEntryWithFilm struct {
	FilmListEntry
	Film Film `db:"film"`
}

type FilmListEntryWithDetails struct {
	FilmListEntryWithFilm
	Poster        media.MediaRelation `db:"poster"`
	TimesWatched  int                 `db:"times_watched"`
	AverageRating *float64            `db:"average_rating"`
	LastWatched   *time.Time          `db:"last_watched"`
}

type FilmListEntryWithPoster struct {
	ID       int                 `db:"id"`
	Position int                 `db:"position"`
	Film     Film                `db:"film"`
	Poster   media.MediaRelation `db:"poster"`
}

type FilmSearchResult struct {
	ID            int      `db:"id"`
	Title         string   `db:"title"`
	Path          string   `db:"path"`
	PosterPath    *string  `db:"poster_path"`
	TimesWatched  int      `db:"times_watched"`
	AverageRating *float64 `db:"average_rating"`
}

type FilmRatingDistribution struct {
	Rating int `db:"rating"`
	Count  int `db:"count"`
}
