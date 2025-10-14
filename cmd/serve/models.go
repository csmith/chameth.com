package main

import "time"

type Poem struct {
	Slug      string    `db:"slug"`
	Title     string    `db:"title"`
	Poem      string    `db:"poem"`
	Notes     string    `db:"notes"`
	Published time.Time `db:"published"`
	Modified  time.Time `db:"modified"`
}

type Snippet struct {
	Slug    string `db:"slug"`
	Title   string `db:"title"`
	Topic   string `db:"topic"`
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
}

type Media struct {
	ID               int    `db:"id"`
	ContentType      string `db:"content_type"`
	OriginalFilename string `db:"original_filename"`
	Data             []byte `db:"data"`
}

type MediaRelation struct {
	Slug        string  `db:"slug"`
	MediaID     int     `db:"media_id"`
	Description *string `db:"description"`
	Caption     *string `db:"caption"`
	Role        string  `db:"role"`
	EntityType  string  `db:"entity_type"`
	EntityID    int     `db:"entity_id"`
}

type Print struct {
	ID          int    `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
}

type PrintLink struct {
	ID      int    `db:"id"`
	PrintID int    `db:"print_id"`
	Name    string `db:"name"`
	Address string `db:"address"`
}
