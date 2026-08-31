package posts

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

func relatedPostsByID(ctx context.Context, postID int, limit int) ([]PostMetadata, error) {
	return db.Select[PostMetadata](ctx, `
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

func recentPostsBySimilarityScore(ctx context.Context, likePaths, unlikePaths []string, unlikeCoeff, maxLikeDist, minScore, minUnlikeDist float64, limit int) ([]Post, error) {
	return db.Select[Post](ctx, `
		WITH likes AS (
			SELECT embedding
			FROM posts
			WHERE published = true
			  AND embedding IS NOT NULL
			  AND path = ANY($1)
		),
		unlikes AS (
			SELECT embedding
			FROM posts
			WHERE published = true
			  AND embedding IS NOT NULL
			  AND path = ANY($2)
		),
		scored AS (
			SELECT id, path, title, date, format, content,
			       (SELECT min(p.embedding <=> l.embedding) FROM likes l) AS like_dist,
			       (SELECT min(p.embedding <=> u.embedding) FROM unlikes u) AS unlike_dist
			FROM posts p
			WHERE published = true
			  AND embedding IS NOT NULL
		)
		SELECT id, path, title, date, format, content
		FROM scored
		WHERE (
			like_dist <= $3
			AND (1 - like_dist) - $4 * COALESCE(1 - unlike_dist, 0) >= $5
		) OR (
			like_dist IS NULL
			AND unlike_dist >= $6
		)
		ORDER BY date DESC
		LIMIT $7
	`, likePaths, unlikePaths, maxLikeDist, unlikeCoeff, minScore, minUnlikeDist, limit)
}
