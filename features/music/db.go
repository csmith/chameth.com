package music

import (
	"context"
	"fmt"

	"chameth.com/chameth.com/db"
)

// rehostedCoverPaths returns the paths of all album and artist art already
// in the media library.
func rehostedCoverPaths(ctx context.Context) (map[string]bool, error) {
	paths, err := db.Select[string](ctx, `
		SELECT path
		FROM media_relations
		WHERE entity_type IN ('album', 'artist')
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list rehosted covers: %w", err)
	}

	set := make(map[string]bool, len(paths))
	for _, p := range paths {
		set[p] = true
	}
	return set, nil
}
