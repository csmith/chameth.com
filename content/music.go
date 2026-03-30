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
	"time"

	"chameth.com/chameth.com/db"
	"chameth.com/chameth.com/external/subsonic"
	"golang.org/x/image/draw"
)

var (
	subsonicBaseUrl  = flag.String("subsonic-base-url", "", "Base URL for the Subsonic API")
	subsonicUsername = flag.String("subsonic-username", "", "Username for the Subsonic API")
	subsonicPassword = flag.String("subsonic-password", "", "Password for the Subsonic API")
)

func RunMusicImport(ctx context.Context, client *http.Client) {
	if *subsonicBaseUrl == "" {
		return
	}

	catalogTicker := time.NewTicker(6 * time.Hour)
	defer catalogTicker.Stop()
	playsTicker := time.NewTicker(1 * time.Minute)
	defer playsTicker.Stop()

	importMusicCatalog(ctx, client)
	importMusicPlays(ctx, client)

	for {
		select {
		case <-ctx.Done():
			return
		case <-catalogTicker.C:
			importMusicCatalog(ctx, client)
		case <-playsTicker.C:
			importMusicPlays(ctx, client)
		}
	}
}

func newSubsonicClient(client *http.Client) *subsonic.Client {
	return subsonic.NewClient(client, *subsonicBaseUrl, *subsonicUsername, *subsonicPassword)
}

func importMusicCatalog(ctx context.Context, client *http.Client) {
	sc := newSubsonicClient(client)

	resp, err := sc.GetArtists()
	if err != nil {
		slog.Error("Failed to get artists", "error", err)
		return
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
				if err := saveArtistImage(ctx, client, id, artist.Name, artist.ArtistImageURL); err != nil {
					slog.Error("Failed to download artist image", "error", err, "name", artist.Name)
				}
			}
		}
	}

	if err := importMusicAlbums(ctx, client, sc); err != nil {
		slog.Error("Failed to import albums", "error", err)
	}

	if err := importMusicTracks(ctx, sc); err != nil {
		slog.Error("Failed to import tracks", "error", err)
	}

	resolveUnmatchedPlays(ctx)
}

func importMusicPlays(ctx context.Context, client *http.Client) {
	sc := newSubsonicClient(client)
	if err := importPlays(ctx, sc); err != nil {
		slog.Error("Failed to import plays", "error", err)
	}
}

func importMusicAlbums(ctx context.Context, client *http.Client, sc *subsonic.Client) error {
	const pageSize = 100
	offset := 0

	for {
		resp, err := sc.GetAlbumList("alphabeticalByName", pageSize, offset)
		if err != nil {
			return err
		}
		if len(resp.Albums) == 0 {
			break
		}

		for _, album := range resp.Albums {
			if album.MusicBrainzID == "" {
				continue
			}

			var artistID *int
			if len(album.AlbumArtists) > 0 {
				id, err := db.GetMusicArtistBySubsonicID(ctx, album.AlbumArtists[0].ID)
				if err != nil {
					slog.Error("Failed to find artist for album", "error", err, "album", album.Title)
				} else {
					artistID = &id
				}
			}

			var year *int
			if album.Year != 0 {
				year = &album.Year
			}

			id, err := db.UpsertMusicAlbum(ctx, db.MusicAlbum{
				MusicBrainzID: album.MusicBrainzID,
				SubsonicID:    album.ID,
				Name:          album.Title,
				SortName:      album.SortName,
				Year:          year,
				ArtistID:      artistID,
			})
			if err != nil {
				slog.Error("Failed to upsert album", "error", err, "name", album.Title)
				continue
			}

			if album.CoverArt != "" {
				if err := saveAlbumCover(ctx, client, id, album.Title, sc.CoverArtURL(album.CoverArt)); err != nil {
					slog.Error("Failed to download album cover", "error", err, "name", album.Title)
				}
			}
		}

		if len(resp.Albums) < pageSize {
			break
		}
		offset += pageSize
	}

	return nil
}

func importMusicTracks(ctx context.Context, sc *subsonic.Client) error {
	albums, err := db.GetAlbumsWithoutTracks(ctx)
	if err != nil {
		return err
	}

	for _, album := range albums {
		detail, err := sc.GetAlbum(album.SubsonicID)
		if err != nil {
			slog.Error("Failed to get album details", "error", err, "name", album.Name)
			continue
		}

		for _, song := range detail.Songs {
			if song.MusicBrainzID == "" {
				continue
			}

			var duration *int
			if song.Duration != 0 {
				duration = &song.Duration
			}

			var discNumber *int
			if song.DiscNumber != 0 {
				discNumber = &song.DiscNumber
			}

			var trackNumber *int
			if song.TrackNumber != 0 {
				trackNumber = &song.TrackNumber
			}

			if _, err := db.UpsertMusicTrack(ctx, db.MusicTrack{
				SubsonicID:    song.ID,
				MusicBrainzID: song.MusicBrainzID,
				AlbumID:       album.ID,
				Name:          song.Title,
				Duration:      duration,
				DiscNumber:    discNumber,
				TrackNumber:   trackNumber,
			}); err != nil {
				slog.Error("Failed to upsert track", "error", err, "name", song.Title)
			}
		}

		slog.Info("Imported tracks for album", "name", album.Name, "count", len(detail.Songs))
	}

	return nil
}

func importPlays(ctx context.Context, sc *subsonic.Client) error {
	mostRecent, err := db.GetMostRecentPlayTime(ctx)
	if err != nil {
		return err
	}

	slog.Info("Importing plays since", "since", mostRecent)

	token, err := sc.LoginNavidrome(ctx)
	if err != nil {
		return fmt.Errorf("failed to login to navidrome: %w", err)
	}

	const pageSize = 100
	offset := 0
	imported := 0

	for {
		plays, err := sc.GetRecentPlays(ctx, token, offset, offset+pageSize)
		if err != nil {
			return err
		}
		if len(plays) == 0 {
			break
		}

		for _, play := range plays {
			if play.Recording == "" {
				continue
			}

			playedAt := play.PlayDate.Truncate(time.Microsecond)

			if playedAt.Before(mostRecent) || playedAt.Equal(mostRecent) {
				slog.Info("Reached previously imported plays", "imported", imported)
				return nil
			}

			trackID, err := db.GetTrackByMusicBrainzID(ctx, play.Recording)
			if err != nil {
				if err := db.InsertUnmatchedMusicPlay(ctx, db.UnmatchedMusicPlay{
					MusicBrainzID: play.Recording,
					Title:         play.Title,
					PlayedAt:      playedAt,
					PlayCount:     play.PlayCount,
				}); err != nil {
					slog.Error("Failed to insert unmatched play", "error", err, "title", play.Title)
				}
				continue
			}

			if err := db.InsertMusicPlay(ctx, db.MusicPlay{
				TrackID:   trackID,
				PlayedAt:  playedAt,
				PlayCount: play.PlayCount,
			}); err != nil {
				slog.Error("Failed to insert play", "error", err, "title", play.Title)
				continue
			}
			imported++
		}

		if len(plays) < pageSize {
			break
		}
		offset += pageSize
	}

	slog.Info("Play import complete", "imported", imported)
	return nil
}

func resolveUnmatchedPlays(ctx context.Context) {
	unmatched, err := db.GetUnmatchedMusicPlays(ctx)
	if err != nil {
		slog.Error("Failed to get unmatched plays", "error", err)
		return
	}
	if len(unmatched) == 0 {
		return
	}

	resolved := 0
	for _, play := range unmatched {
		trackID, err := db.GetTrackByMusicBrainzID(ctx, play.MusicBrainzID)
		if err != nil {
			continue
		}

		if err := db.InsertMusicPlay(ctx, db.MusicPlay{
			TrackID:   trackID,
			PlayedAt:  play.PlayedAt,
			PlayCount: play.PlayCount,
		}); err != nil {
			slog.Error("Failed to insert resolved play", "error", err, "title", play.Title)
			continue
		}

		if err := db.DeleteUnmatchedMusicPlay(ctx, play.ID); err != nil {
			slog.Error("Failed to delete resolved unmatched play", "error", err, "title", play.Title)
			continue
		}
		resolved++
	}

	slog.Info("Resolved unmatched plays", "resolved", resolved, "remaining", len(unmatched)-resolved)
}

func fetchImage(client *http.Client, imageURL string) ([]byte, int, int, error) {
	resp, err := client.Get(strings.Replace(imageURL, "http://", "https://", 1))
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	imgData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to read image: %w", err)
	}

	img, _, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to decode image: %w", err)
	}

	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	const maxShortSide = 500
	if shortSide := min(width, height); shortSide > maxShortSide {
		scale := float64(maxShortSide) / float64(shortSide)
		width = int(float64(width) * scale)
		height = int(float64(height) * scale)
		dst := image.NewRGBA(image.Rect(0, 0, width, height))
		draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
		img = dst
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
		return nil, 0, 0, fmt.Errorf("failed to encode image: %w", err)
	}

	return buf.Bytes(), width, height, nil
}

func saveArtistImage(ctx context.Context, client *http.Client, artistID int, name, imageURL string) error {
	if ok, err := db.HasMediaRelationForEntity(ctx, "artist", artistID, "image"); err != nil {
		return err
	} else if ok {
		return nil
	}

	imgData, width, height, err := fetchImage(client, imageURL)
	if err != nil {
		return err
	}

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

func saveAlbumCover(ctx context.Context, client *http.Client, albumID int, name, imageURL string) error {
	if ok, err := db.HasMediaRelationForEntity(ctx, "album", albumID, "image"); err != nil {
		return err
	} else if ok {
		return nil
	}

	imgData, width, height, err := fetchImage(client, imageURL)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("music-album-%d.jpg", albumID)
	mediaPath := fmt.Sprintf("/music/albums/%d/cover.jpg", albumID)

	mediaID, err := db.CreateMedia(ctx, "image/jpeg", filename, imgData, &width, &height, nil)
	if err != nil {
		return fmt.Errorf("failed to create media: %w", err)
	}

	description := fmt.Sprintf("Cover art for %s", name)
	caption := name
	role := "image"
	if err := db.CreateMediaRelation(ctx, "album", albumID, mediaID, mediaPath, &caption, &description, &role); err != nil {
		return fmt.Errorf("failed to create media relation: %w", err)
	}

	slog.Info("Downloaded album cover", "name", name)
	return nil
}
