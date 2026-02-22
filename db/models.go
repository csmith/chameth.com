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
	Raw       bool   `db:"raw"`
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

type MediaImageVariant struct {
	Path        string `db:"path"`
	ContentType string `db:"content_type"`
	Description string `db:"description"`
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

type FilmListEntryWithPoster struct {
	ID       int           `db:"id"`
	Position int           `db:"position"`
	Film     Film          `db:"film"`
	Poster   MediaRelation `db:"poster"`
}

type VideoGameMetadata struct {
	ID        int    `db:"id"`
	Title     string `db:"title"`
	Published bool   `db:"published"`
	Path      string `db:"path"`
}

type VideoGame struct {
	VideoGameMetadata
	Platform string `db:"platform"`
	Overview string `db:"overview"`
}

type VideoGameReview struct {
	ID               int       `db:"id"`
	VideoGameID      int       `db:"video_game_id"`
	PlayedDate       time.Time `db:"played_date"`
	Rating           int       `db:"rating"`
	Playtime         *int      `db:"playtime"`
	CompletionStatus *string   `db:"completion_status"`
	Notes            string    `db:"notes"`
	Published        bool      `db:"published"`
}

type VideoGameWithReview struct {
	VideoGame
	Review *VideoGameReview
}

type Syndication struct {
	ID          int    `db:"id"`
	Path        string `db:"path"`
	ExternalURL string `db:"external_url"`
	Name        string `db:"name"`
	Published   bool   `db:"published"`
}

type Walk struct {
	ID                  int       `db:"id"`
	ExternalID          string    `db:"external_id"`
	StartDate           time.Time `db:"start_date"`
	EndDate             time.Time `db:"end_date"`
	DurationSeconds     float64   `db:"duration_seconds"`
	DistanceKm          float64   `db:"distance_km"`
	ElevationGainMeters float64   `db:"elevation_gain_meters"`
}

type BoardgameGame struct {
	ID    string `db:"id"`
	BggID int    `db:"bgg_id"`
	Name  string `db:"name"`
	Year  int    `db:"year"`
	Status string `db:"status"`
}

type BoardgamePlay struct {
	ID     string    `db:"id"`
	GameID string    `db:"game_id"`
	Date   time.Time `db:"date"`
}

type FilmSearchResult struct {
	ID            int      `db:"id"`
	Title         string   `db:"title"`
	Path          string   `db:"path"`
	PosterPath    *string  `db:"poster_path"`
	TimesWatched  int      `db:"times_watched"`
	AverageRating *float64 `db:"average_rating"`
}
