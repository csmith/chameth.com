package char

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"time"

	"chameth.com/chameth.com/db"
	"chameth.com/chameth.com/external/ogrestream"
	"chameth.com/chameth.com/features/media"
)

// ensurePortrait rehosts the character's portrait and returns the local
// path to serve it from. The live portrait (no past date requested) keeps
// the path it has always been served from and is updated in place; a
// historical portrait is stored once at a path derived from its content
// hash, so dated renders never disturb the live image.
func ensurePortrait(ctx context.Context, client *ogrestream.Client, c *ogrestream.Character, at time.Time) (string, error) {
	if c.Portrait == nil || c.Portrait.Path == "" {
		return "", fmt.Errorf("character has no portrait")
	}

	if !at.IsZero() && at.Before(time.Now()) {
		if c.Portrait.Sha256 == "" {
			return "", fmt.Errorf("portrait is missing its content hash")
		}
		return ensureHistoricalPortrait(ctx, client, c)
	}
	return ensureLivePortrait(ctx, client, c)
}

// livePortraitPath is the path a character's up-to-date portrait is
// served from — the same URL it has always used.
func livePortraitPath(name string) string {
	return fmt.Sprintf("/wow/characters/%s.png", name)
}

// historicalPortraitPath is the write-once path for a historical
// portrait, keyed by the portrait's SHA-256.
func historicalPortraitPath(sha256 string) string {
	return fmt.Sprintf("/wow/characters/%s.png", sha256)
}

// ensureLivePortrait updates the character's live portrait in place,
// storing it first if it doesn't exist yet.
func ensureLivePortrait(ctx context.Context, client *ogrestream.Client, c *ogrestream.Character) (string, error) {
	path := livePortraitPath(c.Profile.Name)
	filename := fmt.Sprintf("%s.png", c.Profile.Name)

	data, contentType, err := client.Image(ctx, c.Portrait.Path)
	if err != nil {
		return "", fmt.Errorf("failed to download portrait: %w", err)
	}

	mediaID, err := mediaIDAtPath(ctx, path)
	if errors.Is(err, sql.ErrNoRows) {
		if err := storePortrait(ctx, c.BlizzardID, path, filename, c.Profile.Name, "render", contentType, data); err != nil {
			return "", err
		}
		return path, nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to look up portrait: %w", err)
	}

	width, height := imageDimensions(data)
	if err := media.UpdateMedia(ctx, mediaID, contentType, filename, data, width, height); err != nil {
		return "", fmt.Errorf("failed to update portrait: %w", err)
	}
	return path, nil
}

// ensureHistoricalPortrait stores a historical portrait once at its
// content-addressed path; an existing copy is left alone.
func ensureHistoricalPortrait(ctx context.Context, client *ogrestream.Client, c *ogrestream.Character) (string, error) {
	path := historicalPortraitPath(c.Portrait.Sha256)

	_, err := mediaIDAtPath(ctx, path)
	if err == nil {
		return path, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("failed to look up portrait: %w", err)
	}

	data, contentType, err := client.Image(ctx, c.Portrait.Path)
	if err != nil {
		return "", fmt.Errorf("failed to download portrait: %w", err)
	}

	if err := storePortrait(ctx, c.BlizzardID, path, c.Portrait.Sha256+".png", c.Profile.Name, "snapshot", contentType, data); err != nil {
		return "", err
	}
	return path, nil
}

// mediaIDAtPath returns the id of the media stored at path, or
// sql.ErrNoRows when nothing is stored there yet.
func mediaIDAtPath(ctx context.Context, path string) (int, error) {
	return db.Get[int](ctx, `
		SELECT m.id
		FROM media m
		JOIN media_relations mr ON m.id = mr.media_id
		WHERE mr.path = $1
	`, path)
}

// imageDimensions returns the downloaded image's width and height, or
// nils when its format can't be decoded.
func imageDimensions(data []byte) (width, height *int) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, nil
	}
	return &cfg.Width, &cfg.Height
}

// storePortrait stores an unmodified copy of a downloaded portrait at
// path. The path is unique per portrait, so a render that loses a race
// against a concurrent render storing the same image discards its
// duplicate copy; both converge on the same path either way.
func storePortrait(ctx context.Context, blizzardID int, path, filename, caption, role, contentType string, data []byte) error {
	width, height := imageDimensions(data)

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var mediaID int
	err = tx.QueryRow(`
		INSERT INTO media (content_type, original_filename, data, width, height)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, contentType, filename, data, width, height).Scan(&mediaID)
	if err != nil {
		return fmt.Errorf("failed to create media: %w", err)
	}

	res, err := tx.Exec(`
		INSERT INTO media_relations (path, media_id, caption, description, role, entity_type, entity_id)
		VALUES ($1, $2, $3, NULL, $4, 'wow_character', $5)
		ON CONFLICT (path) DO NOTHING
	`, path, mediaID, caption, role, blizzardID)
	if err != nil {
		return fmt.Errorf("failed to create media relation: %w", err)
	}

	if rows, _ := res.RowsAffected(); rows == 0 {
		// A concurrent render stored this portrait first; the path is
		// already occupied by their identical copy, so just discard ours.
		if err = tx.Rollback(); err != nil {
			return fmt.Errorf("failed to roll back duplicate media: %w", err)
		}
		return nil
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
