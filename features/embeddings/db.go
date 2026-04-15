package embeddings

import (
	"context"
	"fmt"

	"chameth.com/chameth.com/db"
)

func updatePostEmbedding(ctx context.Context, path string, embedding any) error {
	_, err := db.Exec(ctx, "UPDATE posts SET embedding = $1 WHERE path = $2", embedding, path)
	if err != nil {
		return fmt.Errorf("failed to update embedding for post %s: %w", path, err)
	}
	return nil
}

func postPathsWithoutEmbeddings(ctx context.Context) ([]string, error) {
	return db.Select[string](ctx, "SELECT path FROM posts WHERE embedding IS NULL AND published = true ORDER BY date DESC")
}

func relatedPostsByID(ctx context.Context, postID int, limit int) ([]db.PostMetadata, error) {
	return db.Select[db.PostMetadata](ctx, `
		SELECT id, path, title, date, format, published
		FROM posts
		WHERE id != $1
		  AND published = true
		  AND embedding IS NOT NULL
		  AND (SELECT embedding FROM posts WHERE id = $1) IS NOT NULL
		ORDER BY embedding <=> (SELECT embedding FROM posts WHERE id = $1)
		LIMIT $2
	`, postID, limit)
}
