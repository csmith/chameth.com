package db

import (
	"context"
	"fmt"

	"chameth.com/chameth.com/metrics"
)

// GetStaticPageByPath returns a static page for the given path.
// It handles cases where the path may or may not have a trailing slash.
// Returns nil if no static page is found with that path.
func GetStaticPageByPath(ctx context.Context, path string) (*StaticPage, error) {
	metrics.LogQuery(ctx)
	var page StaticPage
	err := db.GetContext(ctx, &page, "SELECT id, path, title, content, raw FROM staticpages WHERE path = $1 OR path = $2", path, path+"/")
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// GetStaticPageByID returns a static page for the given ID.
func GetStaticPageByID(ctx context.Context, id int) (*StaticPage, error) {
	metrics.LogQuery(ctx)
	var page StaticPage
	err := db.GetContext(ctx, &page, "SELECT id, path, title, content, published, raw, sitemap_frequency, sitemap_priority FROM staticpages WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// GetAllStaticPages returns all published static pages without their content.
func GetAllStaticPages(ctx context.Context) ([]StaticPageMetadata, error) {
	metrics.LogQuery(ctx)
	var pages []StaticPageMetadata
	err := db.SelectContext(ctx, &pages, "SELECT id, path, title, published, raw, sitemap_frequency, sitemap_priority FROM staticpages WHERE published = true ORDER BY title ASC")
	if err != nil {
		return nil, err
	}
	return pages, nil
}

// GetDraftStaticPages returns all unpublished static pages without their content.
func GetDraftStaticPages(ctx context.Context) ([]StaticPageMetadata, error) {
	metrics.LogQuery(ctx)
	var pages []StaticPageMetadata
	err := db.SelectContext(ctx, &pages, "SELECT id, path, title, published, raw, sitemap_frequency, sitemap_priority FROM staticpages WHERE published = false ORDER BY title ASC")
	if err != nil {
		return nil, err
	}
	return pages, nil
}

// CreateStaticPage creates a new unpublished static page in the database and returns its ID.
func CreateStaticPage(ctx context.Context, path, title string) (int, error) {
	metrics.LogQuery(ctx)
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

// GetSitemapStaticPages returns all published static pages that have sitemap fields set.
func GetSitemapStaticPages(ctx context.Context) ([]StaticPageMetadata, error) {
	metrics.LogQuery(ctx)
	var pages []StaticPageMetadata
	err := db.SelectContext(ctx, &pages, "SELECT id, path, title, published, raw, sitemap_frequency, sitemap_priority FROM staticpages WHERE published = true AND sitemap_frequency IS NOT NULL AND sitemap_priority IS NOT NULL ORDER BY path ASC")
	if err != nil {
		return nil, err
	}
	return pages, nil
}

// UpdateStaticPage updates a static page in the database.
func UpdateStaticPage(ctx context.Context, id int, path, title, content string, published, raw bool, sitemapFrequency *string, sitemapPriority *float64) error {
	metrics.LogQuery(ctx)
	_, err := db.ExecContext(ctx, `
		UPDATE staticpages
		SET path = $1, title = $2, content = $3, published = $4, raw = $5, sitemap_frequency = $6, sitemap_priority = $7
		WHERE id = $8
	`, path, title, content, published, raw, sitemapFrequency, sitemapPriority, id)
	if err != nil {
		return fmt.Errorf("failed to update static page: %w", err)
	}
	return nil
}
