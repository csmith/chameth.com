package db

import "time"

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
