package db

import (
	"context"
	"fmt"

	"chameth.com/chameth.com/metrics"
)

func UpsertMusicArtist(ctx context.Context, artist MusicArtist) (int, error) {
	metrics.LogQuery(ctx)

	var id int
	err := db.GetContext(ctx, &id, `
		INSERT INTO music_artists (music_brainz_id, subsonic_id, name, sort_name)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (music_brainz_id)
		DO UPDATE SET
			subsonic_id = EXCLUDED.subsonic_id,
			name = EXCLUDED.name,
			sort_name = EXCLUDED.sort_name
		RETURNING id
	`, artist.MusicBrainzID, artist.SubsonicID, artist.Name, artist.SortName)

	if err != nil {
		return 0, fmt.Errorf("failed to upsert music artist: %w", err)
	}

	return id, nil
}

func GetMusicArtistBySubsonicID(ctx context.Context, subsonicID string) (int, error) {
	metrics.LogQuery(ctx)

	var id int
	err := db.GetContext(ctx, &id, `SELECT id FROM music_artists WHERE subsonic_id = $1`, subsonicID)
	if err != nil {
		return 0, fmt.Errorf("failed to get music artist by subsonic id: %w", err)
	}
	return id, nil
}

func UpsertMusicAlbum(ctx context.Context, album MusicAlbum) (int, error) {
	metrics.LogQuery(ctx)

	var id int
	err := db.GetContext(ctx, &id, `
		INSERT INTO music_albums (music_brainz_id, subsonic_id, name, sort_name, year, artist_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (music_brainz_id)
		DO UPDATE SET
			subsonic_id = EXCLUDED.subsonic_id,
			name = EXCLUDED.name,
			sort_name = EXCLUDED.sort_name,
			year = EXCLUDED.year,
			artist_id = EXCLUDED.artist_id
		RETURNING id
	`, album.MusicBrainzID, album.SubsonicID, album.Name, album.SortName, album.Year, album.ArtistID)

	if err != nil {
		return 0, fmt.Errorf("failed to upsert music album: %w", err)
	}

	return id, nil
}

func GetAlbumsWithoutTracks(ctx context.Context) ([]MusicAlbum, error) {
	metrics.LogQuery(ctx)

	var albums []MusicAlbum
	err := db.SelectContext(ctx, &albums, `
		SELECT a.* FROM music_albums a
		WHERE NOT EXISTS (SELECT 1 FROM music_tracks t WHERE t.album_id = a.id)
		ORDER BY a.sort_name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get albums without tracks: %w", err)
	}
	return albums, nil
}

func UpsertMusicTrack(ctx context.Context, track MusicTrack) (int, error) {
	metrics.LogQuery(ctx)

	var id int
	err := db.GetContext(ctx, &id, `
		INSERT INTO music_tracks (subsonic_id, music_brainz_id, album_id, name, duration, disc_number, track_number)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (subsonic_id)
		DO UPDATE SET
			music_brainz_id = EXCLUDED.music_brainz_id,
			album_id = EXCLUDED.album_id,
			name = EXCLUDED.name,
			duration = EXCLUDED.duration,
			disc_number = EXCLUDED.disc_number,
			track_number = EXCLUDED.track_number
		RETURNING id
	`, track.SubsonicID, track.MusicBrainzID, track.AlbumID, track.Name, track.Duration, track.DiscNumber, track.TrackNumber)

	if err != nil {
		return 0, fmt.Errorf("failed to upsert music track: %w", err)
	}

	return id, nil
}
