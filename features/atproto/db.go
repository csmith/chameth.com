package atproto

import (
	"context"

	"chameth.com/chameth.com/db"
	"chameth.com/chameth.com/features/posts"
)

func unsyndicatedPosts(ctx context.Context) ([]posts.PostMetadata, error) {
	return db.Select[posts.PostMetadata](ctx, `
		SELECT id, path, title, date, format, published
		FROM posts
		WHERE published AND path NOT IN (
			SELECT path FROM syndications WHERE name = 'Bluesky'
		)
	`)
}
