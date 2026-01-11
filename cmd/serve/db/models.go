package db

import "time"

type PoemMetadata struct {
	ID        int       `db:"id"`
	Path      string    `db:"path"`
	Title     string    `db:"title"`
	Date      time.Time `db:"date"`
	Published bool      `db:"published"`
}

type Poem struct {
	PoemMetadata
	Poem  string `db:"poem"`
	Notes string `db:"notes"`
}

type SnippetMetadata struct {
	ID        int    `db:"id"`
	Path      string `db:"path"`
	Title     string `db:"title"`
	Topic     string `db:"topic"`
	Published bool   `db:"published"`
}

type Snippet struct {
	SnippetMetadata
	Content string `db:"content"`
}

type StaticPageMetadata struct {
	ID        int    `db:"id"`
	Path      string `db:"path"`
	Title     string `db:"title"`
	Published bool   `db:"published"`
}

type StaticPage struct {
	StaticPageMetadata
	Content string `db:"content"`
}

type ProjectSection struct {
	ID          int    `db:"id"`
	Name        string `db:"name"`
	Sort        int    `db:"sort"`
	Description string `db:"description"`
}

type Project struct {
	ID          int    `db:"id"`
	Section     int    `db:"section"`
	Name        string `db:"name"`
	Icon        string `db:"icon"`
	Pinned      bool   `db:"pinned"`
	Description string `db:"description"`
	Published   bool   `db:"published"`
}

type MediaMetadata struct {
	ID               int    `db:"id"`
	ContentType      string `db:"content_type"`
	OriginalFilename string `db:"original_filename"`
	Width            *int   `db:"width"`
	Height           *int   `db:"height"`
	ParentMediaID    *int   `db:"parent_media_id"`
}

type Media struct {
	MediaMetadata
	Data []byte `db:"data"`
}

type MediaRelation struct {
	Path        string  `db:"path"`
	MediaID     int     `db:"media_id"`
	Description *string `db:"description"`
	Caption     *string `db:"caption"`
	Role        *string `db:"role"`
	EntityType  string  `db:"entity_type"`
	EntityID    int     `db:"entity_id"`
}

type MediaRelationWithDetails struct {
	MediaRelation
	Media
}

type Print struct {
	ID          int    `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	Published   bool   `db:"published"`
}

type PrintLink struct {
	ID      int    `db:"id"`
	PrintID int    `db:"print_id"`
	Name    string `db:"name"`
	Address string `db:"address"`
}

type PostMetadata struct {
	ID        int       `db:"id"`
	Path      string    `db:"path"`
	Title     string    `db:"title"`
	Date      time.Time `db:"date"`
	Format    string    `db:"format"`
	Published bool      `db:"published"`
}

type Post struct {
	PostMetadata
	Content string `db:"content"`
}

type PasteMetadata struct {
	ID        int       `db:"id"`
	Path      string    `db:"path"`
	Title     string    `db:"title"`
	Language  string    `db:"language"`
	Date      time.Time `db:"date"`
	Published bool      `db:"published"`
}

type Paste struct {
	PasteMetadata
	Content string `db:"content"`
}

type GoImport struct {
	ID        int    `db:"id"`
	Path      string `db:"path"`
	VCS       string `db:"vcs"`
	RepoURL   string `db:"repo_url"`
	Published bool   `db:"published"`
}

// MediaImageVariant represents a media image with its URL and content type
type MediaImageVariant struct {
	Path        string `db:"path"`
	ContentType string `db:"content_type"`
}

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

type FilmReviewWithFilmAndPoster struct {
	FilmReview `db:"filmreview"`
	Film       `db:"film"`
	Poster     MediaRelationWithDetails `db:"poster"`
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
	Poster        MediaRelation `db:"poster"`
	TimesWatched  int           `db:"times_watched"`
	AverageRating *float64      `db:"average_rating"`
	LastWatched   *time.Time    `db:"last_watched"`
}
