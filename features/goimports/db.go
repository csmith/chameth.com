package goimports

import (
	"context"
	"fmt"

	"chameth.com/chameth.com/db"
)

func GetGoImportByPrefix(ctx context.Context, path string) (*GoImport, error) {
	goimport, err := db.Get[GoImport](ctx, `
		SELECT id, path, vcs, repo_url, published
		FROM goimports
		WHERE $1 = path OR $1 || '/' = path OR $1 LIKE path || '%'
		ORDER BY LENGTH(path) DESC
		LIMIT 1
	`, path)
	if err != nil {
		return nil, err
	}
	return &goimport, nil
}

func GetGoImportByID(ctx context.Context, id int) (*GoImport, error) {
	goimport, err := db.Get[GoImport](ctx, "SELECT id, path, vcs, repo_url, published FROM goimports WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &goimport, nil
}

func GetAllGoImports(ctx context.Context) ([]GoImport, error) {
	return db.Select[GoImport](ctx, "SELECT id, path, vcs, repo_url, published FROM goimports WHERE published = true ORDER BY path")
}

func GetDraftGoImports(ctx context.Context) ([]GoImport, error) {
	return db.Select[GoImport](ctx, "SELECT id, path, vcs, repo_url, published FROM goimports WHERE published = false ORDER BY path")
}

func CreateGoImport(ctx context.Context, path, vcs, repoUrl string) (int, error) {
	var id int
	err := db.QueryRow(ctx, `
		INSERT INTO goimports (path, vcs, repo_url, published)
		VALUES ($1, $2, $3, false)
		RETURNING id
	`, path, vcs, repoUrl).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create goimport: %w", err)
	}
	return id, nil
}

func UpdateGoImport(ctx context.Context, id int, path, vcs, repoUrl string, published bool) error {
	_, err := db.Exec(ctx, `
		UPDATE goimports
		SET path = $1, vcs = $2, repo_url = $3, published = $4
		WHERE id = $5
	`, path, vcs, repoUrl, published, id)
	if err != nil {
		return fmt.Errorf("failed to update goimport: %w", err)
	}
	return nil
}
