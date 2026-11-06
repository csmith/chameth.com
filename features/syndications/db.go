package syndications

import (
	"context"
	"fmt"

	"chameth.com/chameth.com/db"
	"chameth.com/chameth.com/templates"
)

func GetSyndicationByID(ctx context.Context, id int) (*Syndication, error) {
	syndication, err := db.Get[Syndication](
		ctx,
		`
		SELECT id, path, external_url, name, published, disposition, rel
		FROM syndications WHERE id = $1
	`,
		id,
	)
	if err != nil {
		return nil, err
	}
	return &syndication, nil
}

func GetAllSyndications(ctx context.Context) ([]Syndication, error) {
	return db.Select[Syndication](
		ctx,
		`
		SELECT id, path, external_url, name, published, disposition, rel
		FROM syndications WHERE published = true ORDER BY id
	`,
	)
}

func GetUnpublishedSyndications(ctx context.Context) ([]Syndication, error) {
	return db.Select[Syndication](
		ctx,
		`
		SELECT id, path, external_url, name, published, disposition, rel
		FROM syndications WHERE published = false ORDER BY id
	`,
	)
}

func GetAllSyndicationsWithUnpublished(ctx context.Context) ([]Syndication, error) {
	return db.Select[Syndication](
		ctx,
		`
		SELECT id, path, external_url, name, published, disposition, rel
		FROM syndications ORDER BY id
	`,
	)
}

func GetSyndicationsByPath(ctx context.Context, path, disposition string) ([]Syndication, error) {
	return db.Select[Syndication](
		ctx,
		`
		SELECT id, path, external_url, name, published, disposition, rel
		FROM syndications
		WHERE path = $1 AND published = true AND disposition = $2
	`,
		path,
		disposition,
	)
}

func GetLinksByPath(ctx context.Context, path string) ([]templates.Link, error) {
	return db.Select[templates.Link](
		ctx,
		`
		SELECT COALESCE(rel, 'alternate syndication') AS rel, external_url AS href
		FROM syndications
		WHERE path = $1 AND published = true AND disposition = 'link'
	`,
		path,
	)
}

func CreateSyndication(
	ctx context.Context, path, externalURL, name string, published bool, disposition string, rel *string,
) (int, error) {
	var id int
	err := db.QueryRow(
		ctx,
		`
		INSERT INTO syndications (path, external_url, name, published, disposition, rel)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`,
		path,
		externalURL,
		name,
		published,
		disposition,
		rel,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create syndication: %w", err)
	}
	return id, nil
}

func UpdateSyndication(
	ctx context.Context, id int, path, externalURL, name string, published bool, disposition string, rel *string,
) error {
	_, err := db.Exec(
		ctx,
		`
		UPDATE syndications
		SET path = $1, external_url = $2, name = $3, published = $4, disposition = $5, rel = $6
		WHERE id = $7
	`,
		path,
		externalURL,
		name,
		published,
		disposition,
		rel,
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to update syndication: %w", err)
	}
	return nil
}

func DeleteSyndication(ctx context.Context, id int) error {
	_, err := db.Exec(ctx, "DELETE FROM syndications WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete syndication: %w", err)
	}
	return nil
}

func createSyndicationIfMissing(ctx context.Context, path, externalURL, name, disposition string, rel *string) error {
	_, err := db.Exec(
		ctx,
		`
		INSERT INTO syndications (path, external_url, name, published, disposition, rel)
		SELECT $1::varchar, $2::varchar, $3, true, $4, $5
		WHERE NOT EXISTS (SELECT 1 FROM syndications WHERE path = $1 AND external_url = $2)
	`,
		path,
		externalURL,
		name,
		disposition,
		rel,
	)
	if err != nil {
		return fmt.Errorf("failed to create syndication: %w", err)
	}
	return nil
}

func getSyndicationsByPathAndRel(ctx context.Context, path, rel string) ([]Syndication, error) {
	return db.Select[Syndication](
		ctx,
		`
		SELECT id, path, external_url, name, published, disposition, rel
		FROM syndications WHERE path = $1 AND rel = $2 ORDER BY id
	`,
		path,
		rel,
	)
}

func getSyndicationsByPathAndName(ctx context.Context, path, name string) ([]Syndication, error) {
	return db.Select[Syndication](
		ctx,
		`
		SELECT id, path, external_url, name, published, disposition, rel
		FROM syndications WHERE path = $1 AND name = $2 ORDER BY id
	`,
		path,
		name,
	)
}

func getUnsyndicatedAtProtoPaths(ctx context.Context) ([]string, error) {
	return db.Select[string](
		ctx,
		`
		SELECT path
		FROM posts
		WHERE published AND (
			path NOT IN (SELECT path FROM syndications WHERE name = 'Bluesky')
			OR path NOT IN (SELECT path FROM syndications WHERE rel = 'site.standard.document')
		)
		ORDER BY date
	`,
	)
}
