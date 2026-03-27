package content

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"chameth.com/chameth.com/db"
	"chameth.com/chameth.com/external/subsonic"
	"golang.org/x/image/draw"
)

var (
	subsonicBaseUrl  = flag.String("subsonic-base-url", "", "Base URL for the Subsonic API")
	subsonicUsername = flag.String("subsonic-username", "", "Username for the Subsonic API")
	subsonicPassword = flag.String("subsonic-password", "", "Password for the Subsonic API")
)

func ImportMusicDetails(ctx context.Context, client *http.Client) error {
	if *subsonicBaseUrl == "" {
		return fmt.Errorf("subsonic not configured")
	}

	sc := subsonic.NewClient(client, *subsonicBaseUrl, *subsonicUsername, *subsonicPassword)
	resp, err := sc.GetArtists()
	if err != nil {
		return err
	}

	for _, idx := range resp.Indexes {
		for _, artist := range idx.Artists {
			if artist.MusicBrainzID == "" {
				continue
			}

			id, err := db.UpsertMusicArtist(ctx, db.MusicArtist{
				MusicBrainzID: artist.MusicBrainzID,
				SubsonicID:    artist.ID,
				Name:          artist.Name,
				SortName:      artist.SortName,
			})
			if err != nil {
				slog.Error("Failed to upsert artist", "error", err, "name", artist.Name)
				continue
			}

			if artist.ArtistImageURL != "" {
				if err := downloadArtistImage(ctx, client, id, artist.Name, artist.ArtistImageURL); err != nil {
					slog.Error("Failed to download artist image", "error", err, "name", artist.Name)
				}
			}
		}
	}

	return nil
}

func downloadArtistImage(ctx context.Context, client *http.Client, artistID int, name, imageURL string) error {
	existing, err := db.GetMediaRelationsForEntity(ctx, "artist", artistID)
	if err != nil {
		return fmt.Errorf("failed to check existing media relations: %w", err)
	}

	for _, rel := range existing {
		if rel.Role != nil && *rel.Role == "image" {
			return nil
		}
	}

	resp, err := client.Get(strings.Replace(imageURL, "http://", "https://", 1))
	if err != nil {
		return fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	imgData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read image: %w", err)
	}

	img, _, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	var buf bytes.Buffer
	const maxShortSide = 500
	shortSide := min(width, height)
	if shortSide > maxShortSide {
		scale := float64(maxShortSide) / float64(shortSide)
		width = int(float64(width) * scale)
		height = int(float64(height) * scale)
		dst := image.NewRGBA(image.Rect(0, 0, width, height))
		draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
		img = dst
	}
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}
	imgData = buf.Bytes()

	filename := fmt.Sprintf("music-artist-%d.jpg", artistID)
	mediaPath := fmt.Sprintf("/music/artists/%d/cover.jpg", artistID)

	mediaID, err := db.CreateMedia(ctx, "image/jpeg", filename, imgData, &width, &height, nil)
	if err != nil {
		return fmt.Errorf("failed to create media: %w", err)
	}

	description := fmt.Sprintf("Image of %s", name)
	caption := name
	role := "image"
	if err := db.CreateMediaRelation(ctx, "artist", artistID, mediaID, mediaPath, &caption, &description, &role); err != nil {
		return fmt.Errorf("failed to create media relation: %w", err)
	}

	slog.Info("Downloaded artist image", "name", name)
	return nil
}
