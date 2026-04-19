package db

import (
	"context"
	"database/sql"
	"errors"
)

func GetMediaByPath(ctx context.Context, path string) (*Media, error) {
	media, err := Get[Media](ctx, `
		SELECT m.id, m.content_type, m.original_filename, m.data
		FROM media m
		JOIN media_relations mr ON m.id = mr.media_id
		WHERE mr.path = $1
	`, path)
	if err != nil {
		return nil, err
	}
	return &media, nil
}

func GetMediaRelationsForEntity(ctx context.Context, entityType string, entityID int) ([]MediaRelationWithDetails, error) {
	return Select[MediaRelationWithDetails](ctx, `
		SELECT
			mr.path, mr.media_id, mr.description, mr.caption, mr.role, mr.entity_type, mr.entity_id,
			m.id, m.content_type, m.original_filename, m.width, m.height, m.parent_media_id
		FROM media_relations mr
		JOIN media m ON mr.media_id = m.id
		WHERE mr.entity_type = $1 AND mr.entity_id = $2
	`, entityType, entityID)
}

func HasMediaRelationForEntity(ctx context.Context, entityType string, entityID int, role string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM media_relations WHERE entity_type = $1 AND entity_id = $2`
	args := []any{entityType, entityID}
	if role != "" {
		query += ` AND role = $3`
		args = append(args, role)
	}
	query += `)`
	return Get[bool](ctx, query, args...)
}

func GetOpenGraphDetailsForEntity(ctx context.Context, entityType string, entityID int) (*MediaRelationWithDetails, error) {
	relation, err := Get[MediaRelationWithDetails](ctx, `
		SELECT
			mr.path, mr.media_id, mr.description, mr.caption, mr.role, mr.entity_type, mr.entity_id,
			m.id, m.content_type, m.original_filename, m.width, m.height, m.parent_media_id, m.data
		FROM media_relations mr
		JOIN media m ON mr.media_id = m.id
		WHERE mr.entity_type = $1 AND mr.entity_id = $2 AND mr.role = 'opengraph'
		LIMIT 1
	`, entityType, entityID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &relation, nil
}

func GetOpenGraphImageForEntity(ctx context.Context, entityType string, entityID int) (string, error) {
	path, err := Get[string](ctx, `
		SELECT mr.path
		FROM media_relations mr
		WHERE mr.entity_type = $1 AND mr.entity_id = $2 AND mr.role = 'opengraph'
		LIMIT 1
	`, entityType, entityID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return path, nil
}

func GetOpenGraphImageVariantsForEntity(ctx context.Context, entityType string, entityID int) ([]MediaImageVariant, error) {
	return Select[MediaImageVariant](ctx, `
		SELECT mr.path, m.content_type, COALESCE(mr.description, '') AS description
		FROM media_relations mr
		JOIN media m ON mr.media_id = m.id
		WHERE mr.entity_type = $1 AND mr.entity_id = $2 AND mr.role = 'opengraph'
		UNION ALL
		SELECT mr2.path, m2.content_type, '' AS description
		FROM media_relations mr
		JOIN media m ON mr.media_id = m.id
		JOIN media m2 ON m2.parent_media_id = m.id
		JOIN media_relations mr2 ON mr2.media_id = m2.id
		WHERE mr.entity_type = $1 AND mr.entity_id = $2 AND mr.role = 'opengraph'
		  AND mr2.entity_type = $1 AND mr2.entity_id = $2
	`, entityType, entityID)
}

func CreateMedia(ctx context.Context, contentType, originalFilename string, data []byte, width, height *int, parentMediaID *int) (int, error) {
	return Get[int](ctx, `
		INSERT INTO media (content_type, original_filename, data, width, height, parent_media_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, contentType, originalFilename, data, width, height, parentMediaID)
}

func UpdateMediaData(ctx context.Context, id int, data []byte, width, height *int) error {
	_, err := Exec(ctx, `
		UPDATE media SET data = $1, width = $2, height = $3 WHERE id = $4
	`, data, width, height, id)
	return err
}

func UpdateMedia(ctx context.Context, id int, contentType, originalFilename string, data []byte, width, height *int) error {
	_, err := Exec(ctx, `
		UPDATE media SET content_type = $1, original_filename = $2, data = $3, width = $4, height = $5 WHERE id = $6
	`, contentType, originalFilename, data, width, height, id)
	return err
}

func GetAllMedia(ctx context.Context) ([]MediaMetadata, error) {
	return Select[MediaMetadata](ctx, `
		SELECT id, content_type, original_filename, width, height, parent_media_id
		FROM media
		ORDER BY id DESC
	`)
}

func GetMediaByID(ctx context.Context, id int) (*Media, error) {
	media, err := Get[Media](ctx, `
		SELECT id, content_type, original_filename, data, width, height, parent_media_id
		FROM media
		WHERE id = $1
	`, id)
	if err != nil {
		return nil, err
	}
	return &media, nil
}

func UpdateMediaRelation(ctx context.Context, entityType string, entityID int, path string, caption, description, role *string) error {
	_, err := Exec(ctx, `
		UPDATE media_relations
		SET caption = $1, description = $2, role = $3
		WHERE entity_type = $4 AND entity_id = $5 AND path = $6
	`, caption, description, role, entityType, entityID, path)
	return err
}

func DeleteMediaRelation(ctx context.Context, entityType string, entityID int, path string) error {
	_, err := Exec(ctx, `
		DELETE FROM media_relations
		WHERE entity_type = $1 AND entity_id = $2 AND path = $3
	`, entityType, entityID, path)
	return err
}

func UpdateMediaRelationVariants(ctx context.Context, entityType string, entityID, parentMediaID int, caption, description *string) error {
	_, err := Exec(ctx, `
		UPDATE media_relations
		SET caption = $1, description = $2
		WHERE entity_type = $3 AND entity_id = $4
		  AND media_id IN (
			SELECT id FROM media WHERE parent_media_id = $5
		  )
	`, caption, description, entityType, entityID, parentMediaID)
	return err
}

func GetAvailableMediaForEntity(ctx context.Context, entityType string, entityID int) ([]MediaMetadata, error) {
	return Select[MediaMetadata](ctx, `
		SELECT id, content_type, original_filename, width, height, parent_media_id
		FROM media
		WHERE id NOT IN (
			SELECT media_id FROM media_relations
			WHERE entity_type = $1 AND entity_id = $2
		)
		ORDER BY id DESC
	`, entityType, entityID)
}

func CreateMediaRelation(ctx context.Context, entityType string, entityID, mediaID int, path string, caption, description, role *string) error {
	_, err := Exec(ctx, `
		INSERT INTO media_relations (path, media_id, caption, description, role, entity_type, entity_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, path, mediaID, caption, description, role, entityType, entityID)
	return err
}

func DeleteMedia(ctx context.Context, id int) error {
	_, err := Exec(ctx, "DELETE FROM media WHERE id = $1", id)
	return err
}
