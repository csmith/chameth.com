package db

import (
	"context"
	"fmt"

	"chameth.com/chameth.com/metrics"
)

// GetSnippetByPath returns a snippet for the given path.
// It handles cases where the path may or may not have a trailing slash.
// Returns nil if no snippet is found with that path.
func GetSnippetByPath(ctx context.Context, path string) (*Snippet, error) {
	metrics.LogQuery(ctx)
	var snippet Snippet
	err := db.GetContext(ctx, &snippet, "SELECT id, path, title, topic, content, published FROM snippets WHERE path = $1 OR path = $2", path, path+"/")
	if err != nil {
		return nil, err
	}
	return &snippet, nil
}

// GetSnippetByID returns a snippet for the given ID.
func GetSnippetByID(ctx context.Context, id int) (*Snippet, error) {
	metrics.LogQuery(ctx)
	var snippet Snippet
	err := db.GetContext(ctx, &snippet, "SELECT id, path, title, topic, content, published FROM snippets WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &snippet, nil
}

// GetAllSnippets returns all published snippets without their content.
func GetAllSnippets(ctx context.Context) ([]SnippetMetadata, error) {
	metrics.LogQuery(ctx)
	var snippets []SnippetMetadata
	err := db.SelectContext(ctx, &snippets, "SELECT id, path, title, topic, published FROM snippets WHERE published = true ORDER BY topic, title")
	if err != nil {
		return nil, err
	}
	return snippets, nil
}

// GetDraftSnippets returns all unpublished snippets without their content.
func GetDraftSnippets(ctx context.Context) ([]SnippetMetadata, error) {
	metrics.LogQuery(ctx)
	var snippets []SnippetMetadata
	err := db.SelectContext(ctx, &snippets, "SELECT id, path, title, topic, published FROM snippets WHERE published = false ORDER BY topic, title")
	if err != nil {
		return nil, err
	}
	return snippets, nil
}

// CreateSnippet creates a new unpublished snippet in the database and returns its ID.
func CreateSnippet(ctx context.Context, path, title string) (int, error) {
	metrics.LogQuery(ctx)
	var id int
	err := db.QueryRowContext(ctx, `
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
func UpdateSnippet(ctx context.Context, id int, path, title, topic, content string, published bool) error {
	metrics.LogQuery(ctx)
	_, err := db.ExecContext(ctx, `
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
func GetAllTopics(ctx context.Context) ([]string, error) {
	metrics.LogQuery(ctx)
	var topics []string
	err := db.SelectContext(ctx, &topics, "SELECT DISTINCT topic FROM snippets WHERE topic != '' ORDER BY topic")
	if err != nil {
		return nil, err
	}
	return topics, nil
}

// GetRecentSnippetsWithContent returns the N most recent snippets with full content.
func GetRecentSnippetsWithContent(ctx context.Context, limit int) ([]Snippet, error) {
	metrics.LogQuery(ctx)
	var snippets []Snippet
	err := db.SelectContext(ctx, &snippets, `
		SELECT id, path, title, topic, content, published
		FROM snippets
		WHERE published = true
		ORDER BY id DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	return snippets, nil
}
