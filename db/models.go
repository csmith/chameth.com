package db

import "time"

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
