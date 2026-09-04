package boardgames

import (
	"context"
	"log/slog"
	"net/http"
)

// EnsureImages makes sure every game with remote art has a local rehosted
// copy, downloading any that are missing, and returns each BGG id's local
// media path (absent when the game has no art, or it could not be
// fetched). Magic Meters serves art resize-ready and write-once, so copies
// are made as-is, once. Per-game failures are logged and skipped rather
// than failing the retrieval that triggered them; the game renders without
// art until the next refresh. Concurrent retrieves of the same missing art
// may both download it — storing is idempotent, so they converge on one
// copy.
func EnsureImages(ctx context.Context, client *http.Client, games []Game) (map[int]string, error) {
	existing, err := rehostedImagePaths(ctx)
	if err != nil {
		return nil, err
	}

	images := make(map[int]string, len(games))
	for _, g := range games {
		if g.ImageURL == "" || g.BggID == nil {
			continue
		}

		path := imagePath(*g.BggID)
		if existing[path] {
			images[*g.BggID] = path
			continue
		}

		if err := rehostImage(ctx, client, g); err != nil {
			// Already logged; the game renders without art until the
			// next refresh.
			continue
		}
		images[*g.BggID] = path
	}
	return images, nil
}

// rehostImage downloads a game's box art from Magic Meters and stores an
// unmodified copy in the media table.
func rehostImage(ctx context.Context, client *http.Client, g Game) error {
	data, contentType, err := fetchImage(ctx, client, g.ID)
	if err != nil {
		slog.Error("Failed to fetch boardgame image", "game", g.Name, "id", g.ID, "error", err)
		return err
	}

	if err := createBoardgameImage(ctx, *g.BggID, g.Name, contentType, data); err != nil {
		slog.Error("Failed to store boardgame image", "game", g.Name, "id", g.ID, "error", err)
		return err
	}

	slog.Info("Rehosted boardgame image", "game", g.Name, "bgg_id", *g.BggID)
	return nil
}
