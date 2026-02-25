package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log/slog"
	"net/http"
	"path"
	"time"

	"chameth.com/chameth.com/db"
	"golang.org/x/image/draw"
)

func ImportBoardgamesHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var data struct {
			Games []struct {
				ID       int    `json:"id"`
				UUID     string `json:"uuid"`
				BggID    int    `json:"bggId"`
				BggName  string `json:"bggName"`
				BggYear  int    `json:"bggYear"`
				URLImage    string `json:"urlImage"`
				IsBaseGame  int    `json:"isBaseGame"`
				Copies   []struct {
					StatusOwned     int `json:"statusOwned"`
					StatusPrevOwned int `json:"statusPrevOwned"`
				} `json:"copies"`
			} `json:"games"`
			Plays []struct {
				UUID      string `json:"uuid"`
				GameRefID int    `json:"gameRefId"`
				PlayDate  string `json:"playDate"`
			} `json:"plays"`
		}

		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			slog.Error("Failed to decode bgstats data", "error", err)
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Build map of bgstats integer game ID → game UUID for play lookups
		gameIDToUUID := make(map[int]string, len(data.Games))

		for _, g := range data.Games {
			gameIDToUUID[g.ID] = g.UUID

			status := "unowned"
			owned := false
			prevOwned := false
			for _, c := range g.Copies {
				if c.StatusOwned == 1 {
					owned = true
				}
				if c.StatusPrevOwned == 1 {
					prevOwned = true
				}
			}
			if owned {
				status = "owned"
			} else if prevOwned {
				status = "sold"
			}

			err := db.UpsertBoardgameGame(r.Context(), db.BoardgameGame{
				ID:     g.UUID,
				BggID:  g.BggID,
				Name:   g.BggName,
				Year:   g.BggYear,
				Status:      status,
				IsExpansion: g.IsBaseGame != 1,
			})
			if err != nil {
				slog.Error("Failed to upsert boardgame game", "uuid", g.UUID, "name", g.BggName, "error", err)
				continue
			}

			if g.URLImage != "" && g.BggID != 0 {
				if err := downloadBoardgameImage(r.Context(), g.BggID, g.BggName, g.URLImage); err != nil {
					slog.Error("Failed to download boardgame image", "bgg_id", g.BggID, "error", err)
				}
			}
		}

		for _, p := range data.Plays {
			gameUUID, ok := gameIDToUUID[p.GameRefID]
			if !ok {
				slog.Error("Failed to find game for play", "play_uuid", p.UUID, "game_ref_id", p.GameRefID)
				continue
			}

			playDate, err := time.Parse("2006-01-02 15:04:05", p.PlayDate)
			if err != nil {
				slog.Error("Failed to parse play date", "play_uuid", p.UUID, "error", err)
				continue
			}

			err = db.UpsertBoardgamePlay(r.Context(), db.BoardgamePlay{
				ID:     p.UUID,
				GameID: gameUUID,
				Date:   playDate,
			})
			if err != nil {
				slog.Error("Failed to upsert boardgame play", "play_uuid", p.UUID, "error", err)
				continue
			}
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func downloadBoardgameImage(ctx context.Context, bggID int, name, imageURL string) error {
	existing, err := db.GetMediaRelationsForEntity(ctx, "boardgame", bggID)
	if err != nil {
		return fmt.Errorf("failed to check existing media relations: %w", err)
	}

	for _, rel := range existing {
		if rel.Role != nil && *rel.Role == "image" {
			return nil
		}
	}

	resp, err := http.Get(imageURL)
	if err != nil {
		return fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	imgData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read image: %w", err)
	}

	img, format, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	const maxShortSide = 500
	shortSide := min(width, height)
	if shortSide > maxShortSide {
		scale := float64(maxShortSide) / float64(shortSide)
		width = int(float64(width) * scale)
		height = int(float64(height) * scale)
		dst := image.NewRGBA(image.Rect(0, 0, width, height))
		draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
		img = dst

		var buf bytes.Buffer
		switch format {
		case "png":
			err = png.Encode(&buf, img)
		default:
			err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
			format = "jpeg"
		}
		if err != nil {
			return fmt.Errorf("failed to encode resized image: %w", err)
		}
		imgData = buf.Bytes()
	}

	contentType := "image/" + format
	ext := path.Ext(imageURL)
	if ext == "" {
		ext = "." + format
	}
	filename := fmt.Sprintf("boardgame-%d%s", bggID, ext)
	mediaPath := fmt.Sprintf("/boardgames/%d/image%s", bggID, ext)

	mediaID, err := db.CreateMedia(ctx, contentType, filename, imgData, &width, &height, nil)
	if err != nil {
		return fmt.Errorf("failed to create media: %w", err)
	}

	description := fmt.Sprintf("Box art of %s", name)
	caption := name
	role := "image"
	if err := db.CreateMediaRelation(ctx, "boardgame", bggID, mediaID, mediaPath, &caption, &description, &role); err != nil {
		return fmt.Errorf("failed to create media relation: %w", err)
	}

	slog.Info("Downloaded boardgame image", "bgg_id", bggID, "name", name)
	return nil
}
