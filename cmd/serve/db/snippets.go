package db

// GetSnippetBySlug returns a snippet for the given slug.
// It handles cases where the slug may or may not have a trailing slash.
// Returns nil if no snippet is found with that slug.
func GetSnippetBySlug(slug string) (*Snippet, error) {
	var snippet Snippet
	err := db.Get(&snippet, "SELECT slug, title, topic, content FROM snippets WHERE slug = $1 OR slug = $2", slug, slug+"/")
	if err != nil {
		return nil, err
	}
	return &snippet, nil
}

// GetAllSnippets returns all snippets without their content.
func GetAllSnippets() ([]Snippet, error) {
	var snippets []Snippet
	err := db.Select(&snippets, "SELECT slug, title, topic FROM snippets ORDER BY topic, title")
	if err != nil {
		return nil, err
	}
	return snippets, nil
}
