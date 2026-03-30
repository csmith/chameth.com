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
	ID               int      `db:"id"`
	Path             string   `db:"path"`
	Title            string   `db:"title"`
	Published        bool     `db:"published"`
	Raw              bool     `db:"raw"`
	SitemapFrequency *string  `db:"sitemap_frequency"`
	SitemapPriority  *float64 `db:"sitemap_priority"`
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

type MonthlyWalkingSpeed struct {
	Month        time.Time `db:"month"`
	AvgSpeedKmh  float64   `db:"avg_speed_kmh"`
}

type BoardgameGame struct {
	ID    string `db:"id"`
	BggID int    `db:"bgg_id"`
	Name  string `db:"name"`
	Year  int    `db:"year"`
	Status      string `db:"status"`
	IsExpansion bool   `db:"is_expansion"`
}

type BoardgameGameWithStats struct {
	Name       string     `db:"name"`
	Year       int        `db:"year"`
	ImagePath  *string    `db:"image_path"`
	PlayCount  int        `db:"play_count"`
	LastPlayed *time.Time `db:"last_played"`
}

type BoardgameGameWithPlayCount struct {
	Name      string  `db:"name"`
	Year      int     `db:"year"`
	ImagePath *string `db:"image_path"`
	PlayCount int     `db:"play_count"`
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

type FilmRatingDistribution struct {
	Rating int `db:"rating"`
	Count  int `db:"count"`
}

type MusicArtist struct {
	ID            int    `db:"id"`
	MusicBrainzID string `db:"music_brainz_id"`
	SubsonicID    string `db:"subsonic_id"`
	Name          string `db:"name"`
	SortName      string `db:"sort_name"`
}

type MusicAlbum struct {
	ID            int    `db:"id"`
	MusicBrainzID string `db:"music_brainz_id"`
	SubsonicID    string `db:"subsonic_id"`
	Name          string `db:"name"`
	SortName      string `db:"sort_name"`
	Year          *int   `db:"year"`
	ArtistID      *int   `db:"artist_id"`
}

type MusicTrack struct {
	ID            int    `db:"id"`
	SubsonicID    string `db:"subsonic_id"`
	MusicBrainzID string `db:"music_brainz_id"`
	AlbumID       int    `db:"album_id"`
	Name          string `db:"name"`
	Duration      *int   `db:"duration"`
	DiscNumber    *int   `db:"disc_number"`
	TrackNumber   *int   `db:"track_number"`
}

type MusicPlay struct {
	ID        int       `db:"id"`
	TrackID   int       `db:"track_id"`
	PlayedAt  time.Time `db:"played_at"`
	PlayCount int       `db:"play_count"`
}

type TopArtist struct {
	Name        string    `db:"name"`
	TrackCount  int       `db:"track_count"`
	AlbumCount  int       `db:"album_count"`
	PlayCount int     `db:"play_count"`
	ImagePath *string `db:"image_path"`
}

type NowPlaying struct {
	ArtistName string    `db:"artist_name"`
	TrackName  string    `db:"track_name"`
	AlbumName  string    `db:"album_name"`
	ImagePath  *string   `db:"image_path"`
	PlayedAt   time.Time `db:"played_at"`
}

type UnmatchedMusicPlay struct {
	ID            int       `db:"id"`
	MusicBrainzID string    `db:"music_brainz_id"`
	Title         string    `db:"title"`
	PlayedAt      time.Time `db:"played_at"`
	PlayCount     int       `db:"play_count"`
}
