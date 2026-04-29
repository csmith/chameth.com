package pastes

import "time"

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
