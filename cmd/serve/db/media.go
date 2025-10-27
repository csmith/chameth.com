package db

import (
	"database/sql"
	"errors"
)

// GetMediaBySlug returns media for the given slug.
// Returns nil if no media is found with that slug.
func GetMediaBySlug(slug string) (*Media, error) {
	var media Media
	err := db.Get(&media, `
		SELECT m.id, m.content_type, m.original_filename, m.data
		FROM media m
		JOIN media_relations mr ON m.id = mr.media_id
		WHERE mr.slug = $1
	`, slug)
	if err != nil {
		return nil, err
	}
	return &media, nil
}

// GetMediaRelationsForEntity returns all media relations for a given entity type and ID.
func GetMediaRelationsForEntity(entityType string, entityID int) ([]MediaRelationWithDetails, error) {
	var relations []MediaRelationWithDetails
	err := db.Select(&relations, `
		SELECT
			mr.slug, mr.media_id, mr.description, mr.caption, mr.role, mr.entity_type, mr.entity_id,
			m.id, m.content_type, m.original_filename, m.width, m.height
		FROM media_relations mr
		JOIN media m ON mr.media_id = m.id
		WHERE mr.entity_type = $1 AND mr.entity_id = $2
	`, entityType, entityID)
	if err != nil {
		return nil, err
	}
	return relations, nil
}

// GetOpenGraphImageForEntity returns the OpenGraph image slug for a given entity, or empty string if none exists.
func GetOpenGraphImageForEntity(entityType string, entityID int) (string, error) {
	var slug string
	err := db.Get(&slug, `
		SELECT mr.slug
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
	return slug, nil
}

// GetOpenGraphImageVariantsForEntity returns the OpenGraph image and all its variants for a given entity.
// Returns the primary OG image first, followed by all variants (webp, avif, etc.)
// Returns empty slice if no OG image exists.
func GetOpenGraphImageVariantsForEntity(entityType string, entityID int) ([]MediaImageVariant, error) {
	var variants []MediaImageVariant
	err := db.Select(&variants, `
		SELECT mr.slug, m.content_type
		FROM media_relations mr
		JOIN media m ON mr.media_id = m.id
		WHERE mr.entity_type = $1 AND mr.entity_id = $2 AND mr.role = 'opengraph'
		UNION ALL
		SELECT mr2.slug, m2.content_type
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
func CreateMedia(contentType, originalFilename string, data []byte, width, height *int, parentMediaID *int) (int, error) {
	var id int
	err := db.Get(&id, `
		INSERT INTO media (content_type, original_filename, data, width, height, parent_media_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, contentType, originalFilename, data, width, height, parentMediaID)
	return id, err
}

// GetAllMedia returns all media items ordered by ID descending (without binary data).
func GetAllMedia() ([]MediaMetadata, error) {
	var media []MediaMetadata
	err := db.Select(&media, `
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
func GetMediaByID(id int) (*Media, error) {
	var media Media
	err := db.Get(&media, `
		SELECT id, content_type, original_filename, data, width, height, parent_media_id
		FROM media
		WHERE id = $1
	`, id)
	if err != nil {
		return nil, err
	}
	return &media, nil
}
