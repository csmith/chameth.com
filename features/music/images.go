package music

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"chameth.com/chameth.com/db"
)

// EnsureAlbumCovers rehosts album art from the service into the media
// library, returning the local path for each album keyed by subsonic id.
// Albums without artwork map to "".
func EnsureAlbumCovers(ctx context.Context, client *http.Client, albums []Album) (map[string]string, error) {
	refs := make([]coverRef, len(albums))
	for i, a := range albums {
		refs[i] = coverRef{
			entityType: "album",
			id:         a.ID,
			subsonicID: a.SubsonicID,
			name:       a.Name,
			coverPath:  a.Cover,
			path:       fmt.Sprintf("/music/albums/%s/cover.jpg", a.SubsonicID),
		}
	}
	return ensureCovers(ctx, client, refs)
}

// EnsureArtistCovers rehosts artist images from the service into the media
// library, returning the local path for each artist keyed by subsonic id.
// Artists without artwork map to "".
func EnsureArtistCovers(ctx context.Context, client *http.Client, artists []Artist) (map[string]string, error) {
	refs := make([]coverRef, len(artists))
	for i, a := range artists {
		refs[i] = coverRef{
			entityType: "artist",
			id:         a.ID,
			subsonicID: a.SubsonicID,
			name:       a.Name,
			coverPath:  a.Cover,
			path:       fmt.Sprintf("/music/artists/%s/cover.jpg", a.SubsonicID),
		}
	}
	return ensureCovers(ctx, client, refs)
}

type coverRef struct {
	entityType string // album or artist
	id         int    // the service's row id
	subsonicID string
	name       string
	coverPath  string // root-relative cover path on the service
	path       string // media library path once rehosted
}

// Image failures are non-fatal; the next shortcode refresh tries them again.
func ensureCovers(ctx context.Context, client *http.Client, covers []coverRef) (map[string]string, error) {
	existing, err := rehostedCoverPaths(ctx)
	if err != nil {
		return nil, err
	}

	paths := make(map[string]string, len(covers))
	for _, c := range covers {
		if c.coverPath == "" || c.subsonicID == "" {
			paths[c.subsonicID] = ""
			continue
		}

		if existing[c.path] {
			paths[c.subsonicID] = c.path
			continue
		}

		if err := rehostCover(ctx, client, c); err != nil {
			continue
		}
		paths[c.subsonicID] = c.path
	}
	return paths, nil
}

// The service serves finished JPEG cover art, so the bytes are stored as-is.
func rehostCover(ctx context.Context, client *http.Client, c coverRef) error {
	data, contentType, err := FetchImage(ctx, client, c.coverPath)
	if err != nil {
		slog.Error("Failed to fetch music artwork", "type", c.entityType, "name", c.name, "error", err)
		return err
	}

	if err := storeCover(ctx, c, contentType, data); err != nil {
		slog.Error("Failed to store music artwork", "type", c.entityType, "name", c.name, "error", err)
		return err
	}

	slog.Info("Rehosted music artwork", "type", c.entityType, "name", c.name)
	return nil
}

func storeCover(ctx context.Context, c coverRef, contentType string, data []byte) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, subtype, _ := strings.Cut(contentType, "/")
	filename := fmt.Sprintf("music-%s-%d.%s", c.entityType, c.id, subtype)
	var mediaID int
	err = tx.QueryRow(`
		INSERT INTO media (content_type, original_filename, data)
		VALUES ($1, $2, $3)
		RETURNING id
	`, contentType, filename, data).Scan(&mediaID)
	if err != nil {
		return fmt.Errorf("failed to create media: %w", err)
	}

	caption := c.name
	description := fmt.Sprintf("Cover art for %s", c.name)
	if c.entityType == "artist" {
		description = fmt.Sprintf("Image of %s", c.name)
	}
	role := "image"
	res, err := tx.Exec(`
		INSERT INTO media_relations (path, media_id, caption, description, role, entity_type, entity_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (path) DO NOTHING
	`, c.path, mediaID, caption, description, role, c.entityType, c.id)
	if err != nil {
		return fmt.Errorf("failed to create media relation: %w", err)
	}

	if rows, _ := res.RowsAffected(); rows == 0 {
		// A concurrent refresh stored this artwork first; the path is
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
