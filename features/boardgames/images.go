package boardgames

import (
	"context"
	"log/slog"
	"net/http"
)

// Image failures are non-fatal; the next shortcode refresh tries them again.
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
			continue
		}
		images[*g.BggID] = path
	}
	return images, nil
}

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
