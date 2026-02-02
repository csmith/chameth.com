package db

import (
	"context"
	"fmt"
)

// GetPoemByPath returns a poem for the given path.
// It handles cases where the path may or may not have a trailing slash.
// Returns nil if no poem is found with that path.
func GetPoemByPath(ctx context.Context, path string) (*Poem, error) {
	var poem Poem
	err := db.GetContext(ctx, &poem, "SELECT id, path, title, poem, notes, date, published FROM poems WHERE path = $1 OR path = $2", path, path+"/")
	if err != nil {
		return nil, err
	}
	return &poem, nil
}

// GetPoemByID returns a poem for the given ID.
func GetPoemByID(ctx context.Context, id int) (*Poem, error) {
	var poem Poem
	err := db.GetContext(ctx, &poem, "SELECT id, path, title, poem, notes, date, published FROM poems WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &poem, nil
}

// GetAllPoems returns all published poems without their content.
func GetAllPoems(ctx context.Context) ([]PoemMetadata, error) {
	var res []PoemMetadata
	err := db.SelectContext(ctx, &res, "SELECT id, path, title, date, published FROM poems WHERE published = true ORDER BY date DESC")
	if err != nil {
		return nil, err
	}
	return res, nil
}

// GetDraftPoems returns all unpublished poems without their content.
func GetDraftPoems(ctx context.Context) ([]PoemMetadata, error) {
	var poems []PoemMetadata
	err := db.SelectContext(ctx, &poems, "SELECT id, path, title, date, published FROM poems WHERE published = false ORDER BY date DESC")
	if err != nil {
		return nil, err
	}
	return poems, nil
}

// CreatePoem creates a new unpublished poem in the database and returns its ID.
func CreatePoem(ctx context.Context, path, title string) (int, error) {
	var id int
	err := db.QueryRowContext(ctx, `
		INSERT INTO poems (path, title, poem, notes, date, published)
		VALUES ($1, $2, '', '', CURRENT_DATE, false)
		RETURNING id
	`, path, title).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create poem: %w", err)
	}
	return id, nil
}

// UpdatePoem updates a poem in the database.
func UpdatePoem(ctx context.Context, id int, path, title, poem, notes, date string, published bool) error {
	_, err := db.ExecContext(ctx, `
		UPDATE poems
		SET path = $1, title = $2, poem = $3, notes = $4, date = $5, published = $6
		WHERE id = $7
	`, path, title, poem, notes, date, published, id)
	if err != nil {
		return fmt.Errorf("failed to update poem: %w", err)
	}
	return nil
}

// GetRecentPoemsWithContent returns the N most recent poems with full content.
func GetRecentPoemsWithContent(ctx context.Context, limit int) ([]Poem, error) {
	var poems []Poem
	err := db.SelectContext(ctx, &poems, `
		SELECT id, path, title, poem, notes, date, published
		FROM poems
		WHERE published = true
		ORDER BY date DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	return poems, nil
}
