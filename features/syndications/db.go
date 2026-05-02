package syndications

import (
	"context"
	"fmt"

	"chameth.com/chameth.com/db"
	"chameth.com/chameth.com/features/posts"
)

func GetSyndicationByID(ctx context.Context, id int) (*Syndication, error) {
	syndication, err := db.Get[Syndication](ctx, "SELECT id, path, external_url, name, published FROM syndications WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &syndication, nil
}

func GetAllSyndications(ctx context.Context) ([]Syndication, error) {
	return db.Select[Syndication](ctx, "SELECT id, path, external_url, name, published FROM syndications WHERE published = true ORDER BY id")
}

func GetUnpublishedSyndications(ctx context.Context) ([]Syndication, error) {
	return db.Select[Syndication](ctx, "SELECT id, path, external_url, name, published FROM syndications WHERE published = false ORDER BY id")
}

func GetAllSyndicationsWithUnpublished(ctx context.Context) ([]Syndication, error) {
	return db.Select[Syndication](ctx, "SELECT id, path, external_url, name, published FROM syndications ORDER BY id")
}

func GetSyndicationsByPath(ctx context.Context, path string) ([]Syndication, error) {
	return db.Select[Syndication](ctx, "SELECT id, path, external_url, name, published FROM syndications WHERE path = $1 AND published = true", path)
}

func CreateSyndication(ctx context.Context, path, externalURL, name string, published bool) (int, error) {
	var id int
	err := db.QueryRow(ctx, `
		INSERT INTO syndications (path, external_url, name, published)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, path, externalURL, name, published).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create syndication: %w", err)
	}
	return id, nil
}

func UpdateSyndication(ctx context.Context, id int, path, externalURL, name string, published bool) error {
	_, err := db.Exec(ctx, `
		UPDATE syndications
		SET path = $1, external_url = $2, name = $3, published = $4
		WHERE id = $5
	`, path, externalURL, name, published, id)
	if err != nil {
		return fmt.Errorf("failed to update syndication: %w", err)
	}
	return nil
}

func DeleteSyndication(ctx context.Context, id int) error {
	_, err := db.Exec(ctx, "DELETE FROM syndications WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete syndication: %w", err)
	}
	return nil
}

func GetUnsyndicatedAtProtoPosts(ctx context.Context) ([]posts.PostMetadata, error) {
	return db.Select[posts.PostMetadata](ctx, `
		SELECT id, path, title, date, format, published
		FROM posts
		WHERE published AND path NOT IN (
			SELECT path FROM syndications WHERE name = 'Bluesky'
		)
	`)
}
