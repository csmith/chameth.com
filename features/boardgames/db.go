package boardgames

import (
	"context"
	"fmt"
	"strings"

	"chameth.com/chameth.com/db"
)

// Preserve the paths used by the old board game importer.
func imagePath(bggID int) string {
	return fmt.Sprintf("/boardgames/%d/image.jpg", bggID)
}

func rehostedImagePaths(ctx context.Context) (map[string]bool, error) {
	paths, err := db.Select[string](ctx, `
		SELECT path
		FROM media_relations
		WHERE entity_type = 'boardgame'
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list boardgame art: %w", err)
	}

	set := make(map[string]bool, len(paths))
	for _, p := range paths {
		set[p] = true
	}
	return set, nil
}

func createBoardgameImage(ctx context.Context, bggID int, name, contentType string, data []byte) error {
	mediaPath := imagePath(bggID)
	_, subtype, _ := strings.Cut(contentType, "/")
	filename := "boardgame-" + fmt.Sprint(bggID) + "." + subtype

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
		INSERT INTO media (content_type, original_filename, data)
		VALUES ($1, $2, $3)
		RETURNING id
	`, contentType, filename, data).Scan(&mediaID)
	if err != nil {
		return fmt.Errorf("failed to create media: %w", err)
	}

	description := fmt.Sprintf("Box art of %s", name)
	caption := name
	role := "image"
	res, err := tx.Exec(`
		INSERT INTO media_relations (path, media_id, caption, description, role, entity_type, entity_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (path) DO NOTHING
	`, mediaPath, mediaID, caption, description, role, "boardgame", bggID)
	if err != nil {
		return fmt.Errorf("failed to create media relation: %w", err)
	}

	if rows, _ := res.RowsAffected(); rows == 0 {
		// A concurrent render stored this game's art first; the path is
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
