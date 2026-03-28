package db

import (
	"context"
	"fmt"
	"time"

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

func GetMostRecentPlayTime(ctx context.Context) (time.Time, error) {
	metrics.LogQuery(ctx)

	var t *time.Time
	err := db.GetContext(ctx, &t, `
		SELECT GREATEST(
			(SELECT MAX(played_at) FROM music_plays),
			(SELECT MAX(played_at) FROM unmatched_music_plays)
		)`)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get most recent play time: %w", err)
	}
	if t == nil {
		return time.Time{}, nil
	}
	return *t, nil
}

func InsertMusicPlay(ctx context.Context, play MusicPlay) error {
	metrics.LogQuery(ctx)

	_, err := db.ExecContext(ctx, `
		INSERT INTO music_plays (play_id, track_id, played_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (play_id) DO NOTHING
	`, play.PlayID, play.TrackID, play.PlayedAt)
	if err != nil {
		return fmt.Errorf("failed to insert music play: %w", err)
	}
	return nil
}

func GetTrackByMusicBrainzID(ctx context.Context, musicBrainzID string) (int, error) {
	metrics.LogQuery(ctx)

	var id int
	err := db.GetContext(ctx, &id, `SELECT id FROM music_tracks WHERE music_brainz_id = $1`, musicBrainzID)
	if err != nil {
		return 0, fmt.Errorf("failed to get track by music brainz id: %w", err)
	}
	return id, nil
}

func InsertUnmatchedMusicPlay(ctx context.Context, play UnmatchedMusicPlay) error {
	metrics.LogQuery(ctx)

	_, err := db.ExecContext(ctx, `
		INSERT INTO unmatched_music_plays (play_id, music_brainz_id, title, played_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (play_id) DO NOTHING
	`, play.PlayID, play.MusicBrainzID, play.Title, play.PlayedAt)
	if err != nil {
		return fmt.Errorf("failed to insert unmatched music play: %w", err)
	}
	return nil
}

func GetUnmatchedMusicPlays(ctx context.Context) ([]UnmatchedMusicPlay, error) {
	metrics.LogQuery(ctx)

	var plays []UnmatchedMusicPlay
	err := db.SelectContext(ctx, &plays, `SELECT * FROM unmatched_music_plays ORDER BY played_at`)
	if err != nil {
		return nil, fmt.Errorf("failed to get unmatched music plays: %w", err)
	}
	return plays, nil
}

func DeleteUnmatchedMusicPlay(ctx context.Context, id int) error {
	metrics.LogQuery(ctx)

	_, err := db.ExecContext(ctx, `DELETE FROM unmatched_music_plays WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete unmatched music play: %w", err)
	}
	return nil
}
