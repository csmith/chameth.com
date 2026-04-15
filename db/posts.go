package db

import (
	"context"
	"fmt"
)

func GetAllPosts(ctx context.Context) ([]PostMetadata, error) {
	return Select[PostMetadata](ctx, "SELECT id, path, title, date, format, published FROM posts WHERE published = true ORDER BY date DESC")
}

func GetDraftPosts(ctx context.Context) ([]PostMetadata, error) {
	return Select[PostMetadata](ctx, "SELECT id, path, title, date, format, published FROM posts WHERE published = false ORDER BY date DESC")
}

func GetPostByID(ctx context.Context, id int) (*Post, error) {
	post, err := Get[Post](ctx, `
		SELECT id, path, title, content, date, format, published
		FROM posts
		WHERE id = $1
	`, id)
	if err != nil {
		return nil, err
	}

	return &post, nil
}

func CreatePost(ctx context.Context, path, title string) (int, error) {
	var id int
	err := QueryRow(ctx, `
		INSERT INTO posts (path, title, content, date, format, published)
		VALUES ($1, $2, '', CURRENT_DATE, 'long', false)
		RETURNING id
	`, path, title).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create post: %w", err)
	}
	return id, nil
}

func UpdatePost(ctx context.Context, id int, path, title, content, date, format string, published bool) error {
	_, err := Exec(ctx, `
		UPDATE posts
		SET path = $1, title = $2, content = $3, date = $4, format = $5, published = $6
		WHERE id = $7
	`, path, title, content, date, format, published, id)
	if err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}
	return nil
}

func GetPostByPath(ctx context.Context, path string) (*Post, error) {
	post, err := Get[Post](ctx, `
		SELECT id, path, title, content, date, format
		FROM posts
		WHERE path = $1 OR path = $2
	`, path, path+"/")
	if err != nil {
		return nil, err
	}

	return &post, nil
}

func GetRecentPosts(ctx context.Context, limit int) ([]PostMetadata, error) {
	return Select[PostMetadata](ctx, `
		SELECT id, path, title, date, format, published
		FROM posts
		WHERE published = true
		ORDER BY date DESC
		LIMIT $1
	`, limit)
}

func GetRecentPostsWithContent(ctx context.Context, limit int) ([]Post, error) {
	return Select[Post](ctx, `
		SELECT id, path, title, date, format, content
		FROM posts
		WHERE published = true
		ORDER BY date DESC
		LIMIT $1
	`, limit)
}

func GetRecentPostsWithContentByFormat(ctx context.Context, limit int, format string) ([]Post, error) {
	return Select[Post](ctx, `
		SELECT id, path, title, date, format, content
		FROM posts
		WHERE format = $1 AND published = true
		ORDER BY date DESC
		LIMIT $2
	`, format, limit)
}
