package db

import (
	"context"
	"fmt"
)

// GetStaticPageByPath returns a static page for the given path.
// It handles cases where the path may or may not have a trailing slash.
// Returns nil if no static page is found with that path.
func GetStaticPageByPath(ctx context.Context, path string) (*StaticPage, error) {
	var page StaticPage
	err := db.GetContext(ctx, &page, "SELECT id, path, title, content, raw FROM staticpages WHERE path = $1 OR path = $2", path, path+"/")
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// GetStaticPageByID returns a static page for the given ID.
func GetStaticPageByID(ctx context.Context, id int) (*StaticPage, error) {
	var page StaticPage
	err := db.GetContext(ctx, &page, "SELECT id, path, title, content, published, raw FROM staticpages WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// GetAllStaticPages returns all published static pages without their content.
func GetAllStaticPages(ctx context.Context) ([]StaticPageMetadata, error) {
	var pages []StaticPageMetadata
	err := db.SelectContext(ctx, &pages, "SELECT id, path, title, published, raw FROM staticpages WHERE published = true ORDER BY title ASC")
	if err != nil {
		return nil, err
	}
	return pages, nil
}

// GetDraftStaticPages returns all unpublished static pages without their content.
func GetDraftStaticPages(ctx context.Context) ([]StaticPageMetadata, error) {
	var pages []StaticPageMetadata
	err := db.SelectContext(ctx, &pages, "SELECT id, path, title, published, raw FROM staticpages WHERE published = false ORDER BY title ASC")
	if err != nil {
		return nil, err
	}
	return pages, nil
}

// CreateStaticPage creates a new unpublished static page in the database and returns its ID.
func CreateStaticPage(ctx context.Context, path, title string) (int, error) {
	var id int
	err := db.QueryRowContext(ctx, `
		INSERT INTO staticpages (path, title, content, published)
		VALUES ($1, $2, '', false)
		RETURNING id
	`, path, title).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create static page: %w", err)
	}
	return id, nil
}

// UpdateStaticPage updates a static page in the database.
func UpdateStaticPage(ctx context.Context, id int, path, title, content string, published, raw bool) error {
	_, err := db.ExecContext(ctx, `
		UPDATE staticpages
		SET path = $1, title = $2, content = $3, published = $4, raw = $5
		WHERE id = $6
	`, path, title, content, published, raw, id)
	if err != nil {
		return fmt.Errorf("failed to update static page: %w", err)
	}
	return nil
}
