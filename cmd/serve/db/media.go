package db

import (
	"context"
	"database/sql"
	"errors"
)

// GetMediaByPath returns media for the given path.
// Returns nil if no media is found with that path.
func GetMediaByPath(ctx context.Context, path string) (*Media, error) {
	var media Media
	err := db.GetContext(ctx, &media, `
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

// GetMediaRelationsForEntity returns all media relations for a given entity type and ID.
func GetMediaRelationsForEntity(ctx context.Context, entityType string, entityID int) ([]MediaRelationWithDetails, error) {
	var relations []MediaRelationWithDetails
	err := db.SelectContext(ctx, &relations, `
		SELECT
			mr.path, mr.media_id, mr.description, mr.caption, mr.role, mr.entity_type, mr.entity_id,
			m.id, m.content_type, m.original_filename, m.width, m.height, m.parent_media_id
		FROM media_relations mr
		JOIN media m ON mr.media_id = m.id
		WHERE mr.entity_type = $1 AND mr.entity_id = $2
	`, entityType, entityID)
	if err != nil {
		return nil, err
	}
	return relations, nil
}

func GetOpenGraphDetailsForEntity(ctx context.Context, entityType string, entityID int) (*MediaRelationWithDetails, error) {
	var relation MediaRelationWithDetails
	err := db.GetContext(ctx, &relation, `
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

// GetOpenGraphImageForEntity returns the OpenGraph image path for a given entity, or empty string if none exists.
func GetOpenGraphImageForEntity(ctx context.Context, entityType string, entityID int) (string, error) {
	var path string
	err := db.GetContext(ctx, &path, `
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

// GetOpenGraphImageVariantsForEntity returns the OpenGraph image and all its variants for a given entity.
// Returns the primary OG image first, followed by all variants (webp, avif, etc.)
// Returns empty slice if no OG image exists.
func GetOpenGraphImageVariantsForEntity(ctx context.Context, entityType string, entityID int) ([]MediaImageVariant, error) {
	var variants []MediaImageVariant
	err := db.SelectContext(ctx, &variants, `
		SELECT mr.path, m.content_type
		FROM media_relations mr
		JOIN media m ON mr.media_id = m.id
		WHERE mr.entity_type = $1 AND mr.entity_id = $2 AND mr.role = 'opengraph'
		UNION ALL
		SELECT mr2.path, m2.content_type
		FROM media_relations mr
		JOIN media m ON mr.media_id = m.id
		JOIN media m2 ON m2.parent_media_id = m.id
		JOIN media_relations mr2 ON mr2.media_id = m2.id
		WHERE mr.entity_type = $1 AND mr.entity_id = $2 AND mr.role = 'opengraph'
		  AND mr2.entity_type = $1 AND mr2.entity_id = $2
	`, entityType, entityID)
	if err != nil {
		return nil, err
	}
	return variants, nil
}

// CreateMedia creates a new media entry and returns its ID.
func CreateMedia(ctx context.Context, contentType, originalFilename string, data []byte, width, height *int, parentMediaID *int) (int, error) {
	var id int
	err := db.GetContext(ctx, &id, `
		INSERT INTO media (content_type, original_filename, data, width, height, parent_media_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, contentType, originalFilename, data, width, height, parentMediaID)
	return id, err
}

// GetAllMedia returns all media items ordered by ID descending (without binary data).
func GetAllMedia(ctx context.Context) ([]MediaMetadata, error) {
	var media []MediaMetadata
	err := db.SelectContext(ctx, &media, `
		SELECT id, content_type, original_filename, width, height, parent_media_id
		FROM media
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, err
	}
	return media, nil
}

// GetMediaByID returns media by its ID including the binary data.
func GetMediaByID(ctx context.Context, id int) (*Media, error) {
	var media Media
	err := db.GetContext(ctx, &media, `
		SELECT id, content_type, original_filename, data, width, height, parent_media_id
		FROM media
		WHERE id = $1
	`, id)
	if err != nil {
		return nil, err
	}
	return &media, nil
}

// UpdateMediaRelation updates the caption, description, and role for a media relation.
func UpdateMediaRelation(ctx context.Context, entityType string, entityID int, path string, caption, description, role *string) error {
	_, err := db.ExecContext(ctx, `
		UPDATE media_relations
		SET caption = $1, description = $2, role = $3
		WHERE entity_type = $4 AND entity_id = $5 AND path = $6
	`, caption, description, role, entityType, entityID, path)
	return err
}

// DeleteMediaRelation removes a media relation.
func DeleteMediaRelation(ctx context.Context, entityType string, entityID int, path string) error {
	_, err := db.ExecContext(ctx, `
		DELETE FROM media_relations
		WHERE entity_type = $1 AND entity_id = $2 AND path = $3
	`, entityType, entityID, path)
	return err
}

// UpdateMediaRelationVariants updates the caption and description for all variants of a parent media.
func UpdateMediaRelationVariants(ctx context.Context, entityType string, entityID, parentMediaID int, caption, description *string) error {
	_, err := db.ExecContext(ctx, `
		UPDATE media_relations
		SET caption = $1, description = $2
		WHERE entity_type = $3 AND entity_id = $4
		  AND media_id IN (
			SELECT id FROM media WHERE parent_media_id = $5
		  )
	`, caption, description, entityType, entityID, parentMediaID)
	return err
}

// GetAvailableMediaForEntity returns all media not already attached to the given entity, ordered by newest first.
func GetAvailableMediaForEntity(ctx context.Context, entityType string, entityID int) ([]MediaMetadata, error) {
	var media []MediaMetadata
	err := db.SelectContext(ctx, &media, `
		SELECT id, content_type, original_filename, width, height, parent_media_id
		FROM media
		WHERE id NOT IN (
			SELECT media_id FROM media_relations
			WHERE entity_type = $1 AND entity_id = $2
		)
		ORDER BY id DESC
	`, entityType, entityID)
	if err != nil {
		return nil, err
	}
	return media, nil
}

// CreateMediaRelation creates a new media relation.
func CreateMediaRelation(ctx context.Context, entityType string, entityID, mediaID int, path string, caption, description, role *string) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO media_relations (path, media_id, caption, description, role, entity_type, entity_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, path, mediaID, caption, description, role, entityType, entityID)
	return err
}

func DeleteMedia(ctx context.Context, id int) error {
	_, err := db.ExecContext(ctx, "DELETE FROM media WHERE id = $1", id)
	return err
}
