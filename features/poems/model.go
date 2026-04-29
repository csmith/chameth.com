package poems

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
