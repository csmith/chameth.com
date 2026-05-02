package syndications

type Syndication struct {
	ID          int    `db:"id"`
	Path        string `db:"path"`
	ExternalURL string `db:"external_url"`
	Name        string `db:"name"`
	Published   bool   `db:"published"`
}
