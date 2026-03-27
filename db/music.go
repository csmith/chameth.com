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
