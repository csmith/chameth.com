package db

import (
	"fmt"
)

// GetAllPosts returns all posts without their content.
func GetAllPosts() ([]Post, error) {
	var posts []Post
	err := db.Select(&posts, "SELECT id, slug, title, date FROM posts ORDER BY date DESC")
	if err != nil {
		return nil, err
	}
	return posts, nil
}

// GetPostByID returns a post for the given ID.
// Returns an error if no post is found with that ID.
func GetPostByID(id int) (*Post, error) {
	var post Post

	err := db.Get(&post, `
		SELECT id, slug, title, content, date, format
		FROM posts
		WHERE id = $1
	`, id)
	if err != nil {
		return nil, err
	}

	return &post, nil
}

// UpdatePost updates a post in the database.
func UpdatePost(id int, slug, title, content string, date string) error {
	_, err := db.Exec(`
		UPDATE posts
		SET slug = $1, title = $2, content = $3, date = $4
		WHERE id = $5
	`, slug, title, content, date, id)
	if err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}
	return nil
}

// GetPostBySlug returns a post for the given slug.
// It handles cases where the slug may or may not have a trailing slash.
// Returns nil if no post is found with that slug.
func GetPostBySlug(slug string) (*Post, error) {
	var post Post

	err := db.Get(&post, `
		SELECT id, slug, title, content, date, format
		FROM posts
		WHERE slug = $1 OR slug = $2
	`, slug, slug+"/")
	if err != nil {
		return nil, err
	}

	return &post, nil
}

// GetRecentPosts returns the N most recent posts.
func GetRecentPosts(limit int) ([]Post, error) {
	var posts []Post
	err := db.Select(&posts, `
		SELECT id, slug, title, date, format
		FROM posts
		ORDER BY date DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

// GetRecentPostsWithContent returns the N most recent posts with full content.
func GetRecentPostsWithContent(limit int) ([]Post, error) {
	var posts []Post
	err := db.Select(&posts, `
		SELECT id, slug, title, date, format, content
		FROM posts
		ORDER BY date DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

// GetRecentPostsWithContentByFormat returns the N most recent posts with full content filtered by format.
func GetRecentPostsWithContentByFormat(limit int, format string) ([]Post, error) {
	var posts []Post
	err := db.Select(&posts, `
		SELECT id, slug, title, date, format, content
		FROM posts
		WHERE format = $1
		ORDER BY date DESC
		LIMIT $2
	`, format, limit)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

// UpdatePostEmbedding updates the embedding for a post identified by slug.
func UpdatePostEmbedding(slug string, embedding interface{}) error {
	_, err := db.Exec("UPDATE posts SET embedding = $1 WHERE slug = $2", embedding, slug)
	if err != nil {
		return fmt.Errorf("failed to update embedding for post %s: %w", slug, err)
	}
	return nil
}

// GetPostSlugsWithoutEmbeddings returns slugs of all posts that don't have embeddings.
func GetPostSlugsWithoutEmbeddings() ([]string, error) {
	var slugs []string
	err := db.Select(&slugs, "SELECT slug FROM posts WHERE embedding IS NULL ORDER BY date DESC")
	if err != nil {
		return nil, err
	}
	return slugs, nil
}

// GetRelatedPostsByID returns posts that are semantically similar to the given post.
// Returns up to limit posts, ordered by similarity (closest first).
func GetRelatedPostsByID(postID int, limit int) ([]Post, error) {
	var posts []Post
	err := db.Select(&posts, `
		SELECT id, slug, title
		FROM posts
		WHERE id != $1
		  AND embedding IS NOT NULL
		  AND (SELECT embedding FROM posts WHERE id = $1) IS NOT NULL
		ORDER BY embedding <=> (SELECT embedding FROM posts WHERE id = $1)
		LIMIT $2
	`, postID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query related posts: %w", err)
	}
	return posts, nil
}
