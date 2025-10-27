package db

import "fmt"

// GetPoemBySlug returns a poem for the given slug.
// It handles cases where the slug may or may not have a trailing slash.
// Returns nil if no poem is found with that slug.
func GetPoemBySlug(slug string) (*Poem, error) {
	var poem Poem
	err := db.Get(&poem, "SELECT id, slug, title, poem, notes, date, published FROM poems WHERE slug = $1 OR slug = $2", slug, slug+"/")
	if err != nil {
		return nil, err
	}
	return &poem, nil
}

// GetPoemByID returns a poem for the given ID.
func GetPoemByID(id int) (*Poem, error) {
	var poem Poem
	err := db.Get(&poem, "SELECT id, slug, title, poem, notes, date, published FROM poems WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &poem, nil
}

// GetAllPoems returns all published poems without their content.
func GetAllPoems() ([]PoemMetadata, error) {
	var res []PoemMetadata
	err := db.Select(&res, "SELECT id, slug, title, date, published FROM poems WHERE published = true ORDER BY date DESC")
	if err != nil {
		return nil, err
	}
	return res, nil
}

// GetDraftPoems returns all unpublished poems without their content.
func GetDraftPoems() ([]PoemMetadata, error) {
	var poems []PoemMetadata
	err := db.Select(&poems, "SELECT id, slug, title, date, published FROM poems WHERE published = false ORDER BY date DESC")
	if err != nil {
		return nil, err
	}
	return poems, nil
}

// CreatePoem creates a new unpublished poem in the database and returns its ID.
func CreatePoem(slug, title string) (int, error) {
	var id int
	err := db.QueryRow(`
		INSERT INTO poems (slug, title, poem, notes, date, published)
		VALUES ($1, $2, '', '', CURRENT_DATE, false)
		RETURNING id
	`, slug, title).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create poem: %w", err)
	}
	return id, nil
}

// UpdatePoem updates a poem in the database.
func UpdatePoem(id int, slug, title, poem, notes, date string, published bool) error {
	_, err := db.Exec(`
		UPDATE poems
		SET slug = $1, title = $2, poem = $3, notes = $4, date = $5, published = $6
		WHERE id = $7
	`, slug, title, poem, notes, date, published, id)
	if err != nil {
		return fmt.Errorf("failed to update poem: %w", err)
	}
	return nil
}
