package db

import "fmt"

// GetSnippetByPath returns a snippet for the given path.
// It handles cases where the path may or may not have a trailing slash.
// Returns nil if no snippet is found with that path.
func GetSnippetByPath(path string) (*Snippet, error) {
	var snippet Snippet
	err := db.Get(&snippet, "SELECT id, path, title, topic, content, published FROM snippets WHERE path = $1 OR path = $2", path, path+"/")
	if err != nil {
		return nil, err
	}
	return &snippet, nil
}

// GetSnippetByID returns a snippet for the given ID.
func GetSnippetByID(id int) (*Snippet, error) {
	var snippet Snippet
	err := db.Get(&snippet, "SELECT id, path, title, topic, content, published FROM snippets WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &snippet, nil
}

// GetAllSnippets returns all published snippets without their content.
func GetAllSnippets() ([]SnippetMetadata, error) {
	var snippets []SnippetMetadata
	err := db.Select(&snippets, "SELECT id, path, title, topic, published FROM snippets WHERE published = true ORDER BY topic, title")
	if err != nil {
		return nil, err
	}
	return snippets, nil
}

// GetDraftSnippets returns all unpublished snippets without their content.
func GetDraftSnippets() ([]SnippetMetadata, error) {
	var snippets []SnippetMetadata
	err := db.Select(&snippets, "SELECT id, path, title, topic, published FROM snippets WHERE published = false ORDER BY topic, title")
	if err != nil {
		return nil, err
	}
	return snippets, nil
}

// CreateSnippet creates a new unpublished snippet in the database and returns its ID.
func CreateSnippet(path, title string) (int, error) {
	var id int
	err := db.QueryRow(`
		INSERT INTO snippets (path, title, topic, content, published)
		VALUES ($1, $2, '', '', false)
		RETURNING id
	`, path, title).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create snippet: %w", err)
	}
	return id, nil
}

// UpdateSnippet updates a snippet in the database.
func UpdateSnippet(id int, path, title, topic, content string, published bool) error {
	_, err := db.Exec(`
		UPDATE snippets
		SET path = $1, title = $2, topic = $3, content = $4, published = $5
		WHERE id = $6
	`, path, title, topic, content, published, id)
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
