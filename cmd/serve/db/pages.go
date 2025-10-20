package db

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
