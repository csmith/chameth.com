package posts

import "time"

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
