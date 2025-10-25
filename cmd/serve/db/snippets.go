package db

import "fmt"

// GetSnippetBySlug returns a snippet for the given slug.
// It handles cases where the slug may or may not have a trailing slash.
// Returns nil if no snippet is found with that slug.
func GetSnippetBySlug(slug string) (*Snippet, error) {
	var snippet Snippet
	err := db.Get(&snippet, "SELECT id, slug, title, topic, content, published FROM snippets WHERE slug = $1 OR slug = $2", slug, slug+"/")
	if err != nil {
		return nil, err
	}
	return &snippet, nil
}

// GetSnippetByID returns a snippet for the given ID.
func GetSnippetByID(id int) (*Snippet, error) {
	var snippet Snippet
	err := db.Get(&snippet, "SELECT id, slug, title, topic, content, published FROM snippets WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &snippet, nil
}

// GetAllSnippets returns all published snippets without their content.
func GetAllSnippets() ([]Snippet, error) {
	var snippets []Snippet
	err := db.Select(&snippets, "SELECT id, slug, title, topic FROM snippets WHERE published = true ORDER BY topic, title")
	if err != nil {
		return nil, err
	}
	return snippets, nil
}

// GetDraftSnippets returns all unpublished snippets without their content.
func GetDraftSnippets() ([]Snippet, error) {
	var snippets []Snippet
	err := db.Select(&snippets, "SELECT id, slug, title, topic FROM snippets WHERE published = false ORDER BY topic, title")
	if err != nil {
		return nil, err
	}
	return snippets, nil
}

// CreateSnippet creates a new unpublished snippet in the database and returns its ID.
func CreateSnippet(slug, title string) (int, error) {
	var id int
	err := db.QueryRow(`
		INSERT INTO snippets (slug, title, topic, content, published)
		VALUES ($1, $2, '', '', false)
		RETURNING id
	`, slug, title).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create snippet: %w", err)
	}
	return id, nil
}

// UpdateSnippet updates a snippet in the database.
func UpdateSnippet(id int, slug, title, topic, content string, published bool) error {
	_, err := db.Exec(`
		UPDATE snippets
		SET slug = $1, title = $2, topic = $3, content = $4, published = $5
		WHERE id = $6
	`, slug, title, topic, content, published, id)
	if err != nil {
		return fmt.Errorf("failed to update snippet: %w", err)
	}
	return nil
}

// GetAllTopics returns all unique topics from snippets.
func GetAllTopics() ([]string, error) {
	var topics []string
	err := db.Select(&topics, "SELECT DISTINCT topic FROM snippets WHERE topic != '' ORDER BY topic")
	if err != nil {
		return nil, err
	}
	return topics, nil
}
