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
