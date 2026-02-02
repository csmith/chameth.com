package db

import (
	"context"
	"fmt"
)

// GetAllPosts returns all published posts without their content.
func GetAllPosts(ctx context.Context) ([]PostMetadata, error) {
	var posts []PostMetadata
	err := db.SelectContext(ctx, &posts, "SELECT id, path, title, date, format, published FROM posts WHERE published = true ORDER BY date DESC")
	if err != nil {
		return nil, err
	}
	return posts, nil
}

// GetDraftPosts returns all unpublished posts without their content.
func GetDraftPosts(ctx context.Context) ([]PostMetadata, error) {
	var posts []PostMetadata
	err := db.SelectContext(ctx, &posts, "SELECT id, path, title, date, format, published FROM posts WHERE published = false ORDER BY date DESC")
	if err != nil {
		return nil, err
	}
	return posts, nil
}

// GetPostByID returns a post for the given ID.
// Returns an error if no post is found with that ID.
func GetPostByID(ctx context.Context, id int) (*Post, error) {
	var post Post

	err := db.GetContext(ctx, &post, `
		SELECT id, path, title, content, date, format, published
		FROM posts
		WHERE id = $1
	`, id)
	if err != nil {
		return nil, err
	}

	return &post, nil
}

// CreatePost creates a new unpublished post in the database and returns its ID.
func CreatePost(ctx context.Context, path, title string) (int, error) {
	var id int
	err := db.QueryRowContext(ctx, `
		INSERT INTO posts (path, title, content, date, format, published)
		VALUES ($1, $2, '', CURRENT_DATE, 'long', false)
		RETURNING id
	`, path, title).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create post: %w", err)
	}
	return id, nil
}

// UpdatePost updates a post in the database.
func UpdatePost(ctx context.Context, id int, path, title, content, date, format string, published bool) error {
	_, err := db.ExecContext(ctx, `
		UPDATE posts
		SET path = $1, title = $2, content = $3, date = $4, format = $5, published = $6
		WHERE id = $7
	`, path, title, content, date, format, published, id)
	if err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}
	return nil
}

// GetPostByPath returns a post for the given path.
// It handles cases where the path may or may not have a trailing slash.
// Returns nil if no post is found with that path.
func GetPostByPath(ctx context.Context, path string) (*Post, error) {
	var post Post

	err := db.GetContext(ctx, &post, `
		SELECT id, path, title, content, date, format
		FROM posts
		WHERE path = $1 OR path = $2
	`, path, path+"/")
	if err != nil {
		return nil, err
	}

	return &post, nil
}

// GetRecentPosts returns the N most recent posts.
func GetRecentPosts(ctx context.Context, limit int) ([]PostMetadata, error) {
	var posts []PostMetadata
	err := db.SelectContext(ctx, &posts, `
		SELECT id, path, title, date, format, published
		FROM posts
		WHERE published = true
		ORDER BY date DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

// GetRecentPostsWithContent returns the N most recent posts with full content.
func GetRecentPostsWithContent(ctx context.Context, limit int) ([]Post, error) {
	var posts []Post
	err := db.SelectContext(ctx, &posts, `
		SELECT id, path, title, date, format, content
		FROM posts
		WHERE published = true
		ORDER BY date DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

// GetRecentPostsWithContentByFormat returns the N most recent posts with full content filtered by format.
func GetRecentPostsWithContentByFormat(ctx context.Context, limit int, format string) ([]Post, error) {
	var posts []Post
	err := db.SelectContext(ctx, &posts, `
		SELECT id, path, title, date, format, content
		FROM posts
		WHERE format = $1 AND published = true
		ORDER BY date DESC
		LIMIT $2
	`, format, limit)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

// UpdatePostEmbedding updates the embedding for a post identified by path.
func UpdatePostEmbedding(ctx context.Context, path string, embedding interface{}) error {
	_, err := db.ExecContext(ctx, "UPDATE posts SET embedding = $1 WHERE path = $2", embedding, path)
	if err != nil {
		return fmt.Errorf("failed to update embedding for post %s: %w", path, err)
	}
	return nil
}

// GetPostPathsWithoutEmbeddings returns paths of all posts that don't have embeddings.
func GetPostPathsWithoutEmbeddings(ctx context.Context) ([]string, error) {
	var paths []string
	err := db.SelectContext(ctx, &paths, "SELECT path FROM posts WHERE embedding IS NULL AND published = true ORDER BY date DESC")
	if err != nil {
		return nil, err
	}
	return paths, nil
}

// GetRelatedPostsByID returns posts that are semantically similar to the given post.
// Returns up to limit posts, ordered by similarity (closest first).
func GetRelatedPostsByID(ctx context.Context, postID int, limit int) ([]PostMetadata, error) {
	var posts []PostMetadata
	err := db.SelectContext(ctx, &posts, `
		SELECT id, path, title, date, format, published
		FROM posts
		WHERE id != $1
		  AND published = true
		  AND embedding IS NOT NULL
		  AND (SELECT embedding FROM posts WHERE id = $1) IS NOT NULL
		ORDER BY embedding <=> (SELECT embedding FROM posts WHERE id = $1)
		LIMIT $2
	`, postID, limit)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func GetPostsNotSyndicatedToATProto(ctx context.Context) ([]PostMetadata, error) {
	var posts []PostMetadata
	err := db.SelectContext(ctx, &posts, `
		SELECT id, path, title, date, format, published
		FROM posts
		WHERE published AND path NOT IN (
			SELECT path FROM syndications WHERE name = 'Bluesky'
		)
	`)
	if err != nil {
		return nil, err
	}
	return posts, nil
}
