package pages

import (
	"context"
	"fmt"

	"chameth.com/chameth.com/db"
)

func GetStaticPageByPath(ctx context.Context, path string) (*StaticPage, error) {
	page, err := db.Get[StaticPage](ctx, `
		SELECT p.id, p.path, p.title, p.content, p.raw, p.parent_id,
		       parent.path AS parent_path, parent.title AS parent_title
		FROM staticpages p
		LEFT JOIN staticpages parent ON p.parent_id = parent.id
		WHERE p.path = $1 OR p.path = $2
	`, path, path+"/")
	if err != nil {
		return nil, err
	}
	return &page, nil
}

func GetStaticPageByID(ctx context.Context, id int) (*StaticPage, error) {
	page, err := db.Get[StaticPage](ctx, "SELECT id, path, title, content, published, raw, parent_id, sitemap_frequency, sitemap_priority FROM staticpages WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

func GetAllStaticPages(ctx context.Context) ([]StaticPageMetadata, error) {
	return db.Select[StaticPageMetadata](ctx, "SELECT id, path, title, published, raw, parent_id, sitemap_frequency, sitemap_priority FROM staticpages WHERE published = true ORDER BY title ASC")
}

func GetDraftStaticPages(ctx context.Context) ([]StaticPageMetadata, error) {
	return db.Select[StaticPageMetadata](ctx, "SELECT id, path, title, published, raw, parent_id, sitemap_frequency, sitemap_priority FROM staticpages WHERE published = false ORDER BY title ASC")
}

func ListStaticPagesMetadata(ctx context.Context) ([]StaticPageMetadata, error) {
	return db.Select[StaticPageMetadata](ctx, "SELECT id, path, title, published, raw, parent_id, sitemap_frequency, sitemap_priority FROM staticpages ORDER BY title ASC")
}

func CreateStaticPage(ctx context.Context, path, title string) (int, error) {
	var id int
	err := db.QueryRow(ctx, `
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
	return db.Select[StaticPageMetadata](ctx, "SELECT id, path, title, published, raw, sitemap_frequency, sitemap_priority FROM staticpages WHERE published = true AND sitemap_frequency IS NOT NULL AND sitemap_priority IS NOT NULL ORDER BY path ASC")
}

func UpdateStaticPage(ctx context.Context, id int, path, title, content string, published, raw bool, parentID *int, sitemapFrequency *string, sitemapPriority *float64) error {
	_, err := db.Exec(ctx, `
		UPDATE staticpages
		SET path = $1, title = $2, content = $3, published = $4, raw = $5, parent_id = $6, sitemap_frequency = $7, sitemap_priority = $8
		WHERE id = $9
	`, path, title, content, published, raw, parentID, sitemapFrequency, sitemapPriority, id)
	if err != nil {
		return fmt.Errorf("failed to update static page: %w", err)
	}
	return nil
}
