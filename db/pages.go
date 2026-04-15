package db

import (
	"context"
	"fmt"
)

func GetStaticPageByPath(ctx context.Context, path string) (*StaticPage, error) {
	page, err := Get[StaticPage](ctx, "SELECT id, path, title, content, raw FROM staticpages WHERE path = $1 OR path = $2", path, path+"/")
	if err != nil {
		return nil, err
	}
	return &page, nil
}

func GetStaticPageByID(ctx context.Context, id int) (*StaticPage, error) {
	page, err := Get[StaticPage](ctx, "SELECT id, path, title, content, published, raw, sitemap_frequency, sitemap_priority FROM staticpages WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

func GetAllStaticPages(ctx context.Context) ([]StaticPageMetadata, error) {
	return Select[StaticPageMetadata](ctx, "SELECT id, path, title, published, raw, sitemap_frequency, sitemap_priority FROM staticpages WHERE published = true ORDER BY title ASC")
}

func GetDraftStaticPages(ctx context.Context) ([]StaticPageMetadata, error) {
	return Select[StaticPageMetadata](ctx, "SELECT id, path, title, published, raw, sitemap_frequency, sitemap_priority FROM staticpages WHERE published = false ORDER BY title ASC")
}

func CreateStaticPage(ctx context.Context, path, title string) (int, error) {
	var id int
	err := QueryRow(ctx, `
		INSERT INTO staticpages (path, title, content, published)
		VALUES ($1, $2, '', false)
		RETURNING id
	`, path, title).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create static page: %w", err)
	}
	return id, nil
}

func GetSitemapStaticPages(ctx context.Context) ([]StaticPageMetadata, error) {
	return Select[StaticPageMetadata](ctx, "SELECT id, path, title, published, raw, sitemap_frequency, sitemap_priority FROM staticpages WHERE published = true AND sitemap_frequency IS NOT NULL AND sitemap_priority IS NOT NULL ORDER BY path ASC")
}

func UpdateStaticPage(ctx context.Context, id int, path, title, content string, published, raw bool, sitemapFrequency *string, sitemapPriority *float64) error {
	_, err := Exec(ctx, `
		UPDATE staticpages
		SET path = $1, title = $2, content = $3, published = $4, raw = $5, sitemap_frequency = $6, sitemap_priority = $7
		WHERE id = $8
	`, path, title, content, published, raw, sitemapFrequency, sitemapPriority, id)
	if err != nil {
		return fmt.Errorf("failed to update static page: %w", err)
	}
	return nil
}
