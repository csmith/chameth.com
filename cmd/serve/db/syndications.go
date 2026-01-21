package db

import "fmt"

func GetSyndicationByID(id int) (*Syndication, error) {
	var syndication Syndication
	err := db.Get(&syndication, "SELECT id, path, external_url, name, published FROM syndications WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &syndication, nil
}

func GetAllSyndications() ([]Syndication, error) {
	var res []Syndication
	err := db.Select(&res, "SELECT id, path, external_url, name, published FROM syndications WHERE published = true ORDER BY id")
	if err != nil {
		return nil, err
	}
	return res, nil
}

func GetUnpublishedSyndications() ([]Syndication, error) {
	var res []Syndication
	err := db.Select(&res, "SELECT id, path, external_url, name, published FROM syndications WHERE published = false ORDER BY id")
	if err != nil {
		return nil, err
	}
	return res, nil
}

func GetAllSyndicationsWithUnpublished() ([]Syndication, error) {
	var res []Syndication
	err := db.Select(&res, "SELECT id, path, external_url, name, published FROM syndications ORDER BY id")
	if err != nil {
		return nil, err
	}
	return res, nil
}

func GetSyndicationsByPath(path string) ([]Syndication, error) {
	var res []Syndication
	err := db.Select(&res, "SELECT id, path, external_url, name, published FROM syndications WHERE path = $1 AND published = true", path)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func CreateSyndication(path, externalURL, name string) (int, error) {
	var id int
	err := db.QueryRow(`
		INSERT INTO syndications (path, external_url, name, published)
		VALUES ($1, $2, $3, false)
		RETURNING id
	`, path, externalURL, name).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create syndication: %w", err)
	}
	return id, nil
}

func UpdateSyndication(id int, path, externalURL, name string, published bool) error {
	_, err := db.Exec(`
		UPDATE syndications
		SET path = $1, external_url = $2, name = $3, published = $4
		WHERE id = $5
	`, path, externalURL, name, published, id)
	if err != nil {
		return fmt.Errorf("failed to update syndication: %w", err)
	}
	return nil
}

func DeleteSyndication(id int) error {
	_, err := db.Exec("DELETE FROM syndications WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete syndication: %w", err)
	}
	return nil
}
