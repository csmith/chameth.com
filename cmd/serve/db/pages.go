package db

import "fmt"

// GetStaticPageByPath returns a static page for the given path.
// It handles cases where the path may or may not have a trailing slash.
// Returns nil if no static page is found with that path.
func GetStaticPageByPath(path string) (*StaticPage, error) {
	var page StaticPage
	err := db.Get(&page, "SELECT id, path, title, content FROM staticpages WHERE path = $1 OR path = $2", path, path+"/")
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// GetStaticPageByID returns a static page for the given ID.
func GetStaticPageByID(id int) (*StaticPage, error) {
	var page StaticPage
	err := db.Get(&page, "SELECT id, path, title, content, published FROM staticpages WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// GetAllStaticPages returns all published static pages without their content.
func GetAllStaticPages() ([]StaticPageMetadata, error) {
	var pages []StaticPageMetadata
	err := db.Select(&pages, "SELECT id, path, title, published FROM staticpages WHERE published = true ORDER BY title ASC")
	if err != nil {
		return nil, err
	}
	return pages, nil
}

// GetDraftStaticPages returns all unpublished static pages without their content.
func GetDraftStaticPages() ([]StaticPageMetadata, error) {
	var pages []StaticPageMetadata
	err := db.Select(&pages, "SELECT id, path, title, published FROM staticpages WHERE published = false ORDER BY title ASC")
	if err != nil {
		return nil, err
	}
	return pages, nil
}

// CreateStaticPage creates a new unpublished static page in the database and returns its ID.
func CreateStaticPage(path, title string) (int, error) {
	var id int
	err := db.QueryRow(`
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
func UpdateStaticPage(id int, path, title, content string, published bool) error {
	_, err := db.Exec(`
		UPDATE staticpages
		SET path = $1, title = $2, content = $3, published = $4
		WHERE id = $5
	`, path, title, content, published, id)
	if err != nil {
		return fmt.Errorf("failed to update static page: %w", err)
	}
	return nil
}
