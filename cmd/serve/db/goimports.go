package db

import "fmt"

func GetGoImportByPrefix(path string) (*GoImport, error) {
	var goimport GoImport
	err := db.Get(&goimport, `
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

func GetGoImportByID(id int) (*GoImport, error) {
	var goimport GoImport
	err := db.Get(&goimport, "SELECT id, path, vcs, repo_url, published FROM goimports WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &goimport, nil
}

func GetAllGoImports() ([]GoImport, error) {
	var res []GoImport
	err := db.Select(&res, "SELECT id, path, vcs, repo_url, published FROM goimports WHERE published = true ORDER BY path")
	if err != nil {
		return nil, err
	}
	return res, nil
}

func GetDraftGoImports() ([]GoImport, error) {
	var goimports []GoImport
	err := db.Select(&goimports, "SELECT id, path, vcs, repo_url, published FROM goimports WHERE published = false ORDER BY path")
	if err != nil {
		return nil, err
	}
	return goimports, nil
}

func CreateGoImport(path, vcs, repoUrl string) (int, error) {
	var id int
	err := db.QueryRow(`
		INSERT INTO goimports (path, vcs, repo_url, published)
		VALUES ($1, $2, $3, false)
		RETURNING id
	`, path, vcs, repoUrl).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create goimport: %w", err)
	}
	return id, nil
}

func UpdateGoImport(id int, path, vcs, repoUrl string, published bool) error {
	_, err := db.Exec(`
		UPDATE goimports
		SET path = $1, vcs = $2, repo_url = $3, published = $4
		WHERE id = $5
	`, path, vcs, repoUrl, published, id)
	if err != nil {
		return fmt.Errorf("failed to update goimport: %w", err)
	}
	return nil
}
