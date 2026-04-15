package db

import (
	"context"
	"fmt"
)

func GetPasteByPath(ctx context.Context, path string) (*Paste, error) {
	paste, err := Get[Paste](ctx, "SELECT id, path, title, language, date, published, content FROM pastes WHERE path = $1 OR path = $2", path, path+"/")
	if err != nil {
		return nil, err
	}
	return &paste, nil
}

func GetPasteByID(ctx context.Context, id int) (*Paste, error) {
	paste, err := Get[Paste](ctx, "SELECT id, path, title, language, date, published, content FROM pastes WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &paste, nil
}

func GetAllPastes(ctx context.Context) ([]PasteMetadata, error) {
	return Select[PasteMetadata](ctx, "SELECT id, path, title, language, date, published FROM pastes WHERE published = true ORDER BY date DESC")
}

func GetDraftPastes(ctx context.Context) ([]PasteMetadata, error) {
	return Select[PasteMetadata](ctx, "SELECT id, path, title, language, date, published FROM pastes WHERE published = false ORDER BY date DESC")
}

func CreatePaste(ctx context.Context, path, title string) (int, error) {
	var id int
	err := QueryRow(ctx, `
		INSERT INTO pastes (path, title, language, date, published, content)
		VALUES ($1, $2, '', CURRENT_TIMESTAMP, false, '')
		RETURNING id
	`, path, title).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create paste: %w", err)
	}
	return id, nil
}

func UpdatePaste(ctx context.Context, id int, path, title, language, content, date string, published bool) error {
	_, err := Exec(ctx, `
		UPDATE pastes
		SET path = $1, title = $2, language = $3, content = $4, date = $5, published = $6
		WHERE id = $7
	`, path, title, language, content, date, published, id)
	if err != nil {
		return fmt.Errorf("failed to update paste: %w", err)
	}
	return nil
}
