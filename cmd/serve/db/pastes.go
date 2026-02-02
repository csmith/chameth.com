package db

import (
	"context"
	"fmt"
)

// GetPasteByPath returns a paste for the given path.
// It handles cases where the path may or may not have a trailing slash.
// Returns nil if no paste is found with that path.
func GetPasteByPath(ctx context.Context, path string) (*Paste, error) {
	var paste Paste
	err := db.GetContext(ctx, &paste, "SELECT id, path, title, language, date, published, content FROM pastes WHERE path = $1 OR path = $2", path, path+"/")
	if err != nil {
		return nil, err
	}
	return &paste, nil
}

// GetPasteByID returns a paste for the given ID.
func GetPasteByID(ctx context.Context, id int) (*Paste, error) {
	var paste Paste
	err := db.GetContext(ctx, &paste, "SELECT id, path, title, language, date, published, content FROM pastes WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &paste, nil
}

// GetAllPastes returns all published pastes without their content.
func GetAllPastes(ctx context.Context) ([]PasteMetadata, error) {
	var res []PasteMetadata
	err := db.SelectContext(ctx, &res, "SELECT id, path, title, language, date, published FROM pastes WHERE published = true ORDER BY date DESC")
	if err != nil {
		return nil, err
	}
	return res, nil
}

// GetDraftPastes returns all unpublished pastes without their content.
func GetDraftPastes(ctx context.Context) ([]PasteMetadata, error) {
	var pastes []PasteMetadata
	err := db.SelectContext(ctx, &pastes, "SELECT id, path, title, language, date, published FROM pastes WHERE published = false ORDER BY date DESC")
	if err != nil {
		return nil, err
	}
	return pastes, nil
}

// CreatePaste creates a new unpublished paste in the database and returns its ID.
func CreatePaste(ctx context.Context, path, title string) (int, error) {
	var id int
	err := db.QueryRowContext(ctx, `
		INSERT INTO pastes (path, title, language, date, published, content)
		VALUES ($1, $2, '', CURRENT_TIMESTAMP, false, '')
		RETURNING id
	`, path, title).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create paste: %w", err)
	}
	return id, nil
}

// UpdatePaste updates a paste in the database.
func UpdatePaste(ctx context.Context, id int, path, title, language, content, date string, published bool) error {
	_, err := db.ExecContext(ctx, `
		UPDATE pastes
		SET path = $1, title = $2, language = $3, content = $4, date = $5, published = $6
		WHERE id = $7
	`, path, title, language, content, date, published, id)
	if err != nil {
		return fmt.Errorf("failed to update paste: %w", err)
	}
	return nil
}
