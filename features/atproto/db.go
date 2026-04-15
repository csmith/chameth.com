package atproto

import (
	"context"

	"chameth.com/chameth.com/db"
)

func unsyndicatedPosts(ctx context.Context) ([]db.PostMetadata, error) {
	return db.Select[db.PostMetadata](ctx, `
		SELECT id, path, title, date, format, published
		FROM posts
		WHERE published AND path NOT IN (
			SELECT path FROM syndications WHERE name = 'Bluesky'
		)
	`)
}
