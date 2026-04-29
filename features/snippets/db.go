package snippets

import (
	"context"
	"fmt"

	"chameth.com/chameth.com/db"
)

func GetSnippetByPath(ctx context.Context, path string) (*Snippet, error) {
	snippet, err := db.Get[Snippet](ctx, "SELECT id, path, title, topic, content, published FROM snippets WHERE path = $1 OR path = $2", path, path+"/")
	if err != nil {
		return nil, err
	}
	return &snippet, nil
}

func GetSnippetByID(ctx context.Context, id int) (*Snippet, error) {
	snippet, err := db.Get[Snippet](ctx, "SELECT id, path, title, topic, content, published FROM snippets WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &snippet, nil
}

func GetAllSnippets(ctx context.Context) ([]SnippetMetadata, error) {
	return db.Select[SnippetMetadata](ctx, "SELECT id, path, title, topic, published FROM snippets WHERE published = true ORDER BY topic, title")
}

func GetDraftSnippets(ctx context.Context) ([]SnippetMetadata, error) {
	return db.Select[SnippetMetadata](ctx, "SELECT id, path, title, topic, published FROM snippets WHERE published = false ORDER BY topic, title")
}

func CreateSnippet(ctx context.Context, path, title string) (int, error) {
	var id int
	err := db.QueryRow(ctx, `
		INSERT INTO snippets (path, title, topic, content, published)
		VALUES ($1, $2, '', '', false)
		RETURNING id
	`, path, title).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create snippet: %w", err)
	}
	return id, nil
}

func UpdateSnippet(ctx context.Context, id int, path, title, topic, content string, published bool) error {
	_, err := db.Exec(ctx, `
		UPDATE snippets
		SET path = $1, title = $2, topic = $3, content = $4, published = $5
		WHERE id = $6
	`, path, title, topic, content, published, id)
	if err != nil {
		return fmt.Errorf("failed to update snippet: %w", err)
	}
	return nil
}

func GetAllTopics(ctx context.Context) ([]string, error) {
	return db.Select[string](ctx, "SELECT DISTINCT topic FROM snippets WHERE topic != '' ORDER BY topic")
}

func GetRecentSnippetsWithContent(ctx context.Context, limit int) ([]Snippet, error) {
	return db.Select[Snippet](ctx, `
		SELECT id, path, title, topic, content, published
		FROM snippets
		WHERE published = true
		ORDER BY id DESC
		LIMIT $1
	`, limit)
}
