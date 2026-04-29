package poems

import (
	"context"
	"fmt"

	"chameth.com/chameth.com/db"
)

func GetPoemByPath(ctx context.Context, path string) (*Poem, error) {
	poem, err := db.Get[Poem](ctx, "SELECT id, path, title, poem, notes, date, published FROM poems WHERE path = $1 OR path = $2", path, path+"/")
	if err != nil {
		return nil, err
	}
	return &poem, nil
}

func GetPoemByID(ctx context.Context, id int) (*Poem, error) {
	poem, err := db.Get[Poem](ctx, "SELECT id, path, title, poem, notes, date, published FROM poems WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &poem, nil
}

func GetAllPoems(ctx context.Context) ([]PoemMetadata, error) {
	return db.Select[PoemMetadata](ctx, "SELECT id, path, title, date, published FROM poems WHERE published = true ORDER BY date DESC")
}

func GetDraftPoems(ctx context.Context) ([]PoemMetadata, error) {
	return db.Select[PoemMetadata](ctx, "SELECT id, path, title, date, published FROM poems WHERE published = false ORDER BY date DESC")
}

func CreatePoem(ctx context.Context, path, title string) (int, error) {
	var id int
	err := db.QueryRow(ctx, `
		INSERT INTO poems (path, title, poem, notes, date, published)
		VALUES ($1, $2, '', '', CURRENT_DATE, false)
		RETURNING id
	`, path, title).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create poem: %w", err)
	}
	return id, nil
}

func UpdatePoem(ctx context.Context, id int, path, title, poem, notes, date string, published bool) error {
	_, err := db.Exec(ctx, `
		UPDATE poems
		SET path = $1, title = $2, poem = $3, notes = $4, date = $5, published = $6
		WHERE id = $7
	`, path, title, poem, notes, date, published, id)
	if err != nil {
		return fmt.Errorf("failed to update poem: %w", err)
	}
	return nil
}

func GetRecentPoemsWithContent(ctx context.Context, limit int) ([]Poem, error) {
	return db.Select[Poem](ctx, `
		SELECT id, path, title, poem, notes, date, published
		FROM poems
		WHERE published = true
		ORDER BY date DESC
		LIMIT $1
	`, limit)
}
