package snippets

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
