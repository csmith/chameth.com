package music

import (
	"context"
	"fmt"
	"time"

	"chameth.com/chameth.com/db"
)

func upsertArtist(ctx context.Context, artist musicArtist) (int, error) {
	id, err := db.Get[int](ctx, `
		INSERT INTO music_artists (subsonic_id, name, sort_name)
		VALUES ($1, $2, $3)
		ON CONFLICT (subsonic_id)
		DO UPDATE SET
			name = EXCLUDED.name,
			sort_name = EXCLUDED.sort_name
		RETURNING id
	`, artist.SubsonicID, artist.Name, artist.SortName)
	if err != nil {
		return 0, fmt.Errorf("failed to upsert music artist: %w", err)
	}
	return id, nil
}

func artistBySubsonicID(ctx context.Context, subsonicID string) (int, error) {
	id, err := db.Get[int](ctx, `SELECT id FROM music_artists WHERE subsonic_id = $1`, subsonicID)
	if err != nil {
		return 0, fmt.Errorf("failed to get music artist by subsonic id: %w", err)
	}
	return id, nil
}

func upsertAlbum(ctx context.Context, album musicAlbum) (int, error) {
	id, err := db.Get[int](ctx, `
		INSERT INTO music_albums (subsonic_id, name, sort_name, year, artist_id)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (subsonic_id)
		DO UPDATE SET
			name = EXCLUDED.name,
			sort_name = EXCLUDED.sort_name,
			year = EXCLUDED.year,
			artist_id = EXCLUDED.artist_id
		RETURNING id
	`, album.SubsonicID, album.Name, album.SortName, album.Year, album.ArtistID)
	if err != nil {
		return 0, fmt.Errorf("failed to upsert music album: %w", err)
	}
	return id, nil
}

func albumBySubsonicID(ctx context.Context, subsonicID string) (int, error) {
	id, err := db.Get[int](ctx, `SELECT id FROM music_albums WHERE subsonic_id = $1`, subsonicID)
	if err != nil {
		return 0, fmt.Errorf("failed to get album by subsonic id: %w", err)
	}
	return id, nil
}

func albumsWithoutTracks(ctx context.Context) ([]musicAlbum, error) {
	albums, err := db.Select[musicAlbum](ctx, `
		SELECT a.* FROM music_albums a
		WHERE NOT EXISTS (SELECT 1 FROM music_tracks t WHERE t.album_id = a.id)
		ORDER BY a.sort_name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get albums without tracks: %w", err)
	}
	return albums, nil
}

func upsertTrack(ctx context.Context, track musicTrack) (int, error) {
	id, err := db.Get[int](ctx, `
		INSERT INTO music_tracks (subsonic_id, album_id, name, duration, disc_number, track_number)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (subsonic_id)
		DO UPDATE SET
			album_id = EXCLUDED.album_id,
			name = EXCLUDED.name,
			duration = EXCLUDED.duration,
			disc_number = EXCLUDED.disc_number,
			track_number = EXCLUDED.track_number
		RETURNING id
	`, track.SubsonicID, track.AlbumID, track.Name, track.Duration, track.DiscNumber, track.TrackNumber)
	if err != nil {
		return 0, fmt.Errorf("failed to upsert music track: %w", err)
	}
	return id, nil
}

func mostRecentPlayTime(ctx context.Context) (time.Time, error) {
	t, err := db.Get[*time.Time](ctx, `SELECT MAX(played_at) FROM music_plays`)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get most recent play time: %w", err)
	}
	if t == nil {
		return time.Time{}, nil
	}
	return *t, nil
}

func insertPlay(ctx context.Context, play musicPlay) error {
	_, err := db.Exec(ctx, `
		INSERT INTO music_plays (track_id, played_at, play_count)
		VALUES ($1, $2, $3)
	`, play.TrackID, play.PlayedAt, play.PlayCount)
	if err != nil {
		return fmt.Errorf("failed to insert music play: %w", err)
	}
	return nil
}

func trackBySubsonicID(ctx context.Context, subsonicID string) (int, error) {
	id, err := db.Get[int](ctx, `SELECT id FROM music_tracks WHERE subsonic_id = $1`, subsonicID)
	if err != nil {
		return 0, fmt.Errorf("failed to get track by subsonic id: %w", err)
	}
	return id, nil
}
