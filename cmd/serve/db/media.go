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
