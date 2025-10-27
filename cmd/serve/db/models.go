package db

import "time"

type PoemMetadata struct {
	ID        int       `db:"id"`
	Slug      string    `db:"slug"`
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
	Slug      string `db:"slug"`
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
	Slug      string `db:"slug"`
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
	Slug        string  `db:"slug"`
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
	Slug      string    `db:"slug"`
	Title     string    `db:"title"`
	Date      time.Time `db:"date"`
	Format    string    `db:"format"`
	Published bool      `db:"published"`
}

type Post struct {
	PostMetadata
	Content string `db:"content"`
}

// MediaImageVariant represents a media image with its URL and content type
type MediaImageVariant struct {
	Slug        string `db:"slug"`
	ContentType string `db:"content_type"`
}
