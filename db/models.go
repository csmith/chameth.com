package db

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
