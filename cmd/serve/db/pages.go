package db

import "fmt"

// GetStaticPageBySlug returns a static page for the given slug.
// It handles cases where the slug may or may not have a trailing slash.
// Returns nil if no static page is found with that slug.
func GetStaticPageBySlug(slug string) (*StaticPage, error) {
	var page StaticPage
	err := db.Get(&page, "SELECT id, slug, title, content FROM staticpages WHERE slug = $1 OR slug = $2", slug, slug+"/")
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// GetStaticPageByID returns a static page for the given ID.
func GetStaticPageByID(id int) (*StaticPage, error) {
	var page StaticPage
	err := db.Get(&page, "SELECT id, slug, title, content, published FROM staticpages WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// GetAllStaticPages returns all published static pages without their content.
func GetAllStaticPages() ([]StaticPageMetadata, error) {
	var pages []StaticPageMetadata
	err := db.Select(&pages, "SELECT id, slug, title, published FROM staticpages WHERE published = true ORDER BY title ASC")
	if err != nil {
		return nil, err
	}
	return pages, nil
}

// GetDraftStaticPages returns all unpublished static pages without their content.
func GetDraftStaticPages() ([]StaticPageMetadata, error) {
	var pages []StaticPageMetadata
	err := db.Select(&pages, "SELECT id, slug, title, published FROM staticpages WHERE published = false ORDER BY title ASC")
	if err != nil {
		return nil, err
	}
	return pages, nil
}

// CreateStaticPage creates a new unpublished static page in the database and returns its ID.
func CreateStaticPage(slug, title string) (int, error) {
	var id int
	err := db.QueryRow(`
		INSERT INTO staticpages (slug, title, content, published)
		VALUES ($1, $2, '', false)
		RETURNING id
	`, slug, title).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create static page: %w", err)
	}
	return id, nil
}

// UpdateStaticPage updates a static page in the database.
func UpdateStaticPage(id int, slug, title, content string, published bool) error {
	_, err := db.Exec(`
		UPDATE staticpages
		SET slug = $1, title = $2, content = $3, published = $4
		WHERE id = $5
	`, slug, title, content, published, id)
	if err != nil {
		return fmt.Errorf("failed to update static page: %w", err)
	}
	return nil
}
