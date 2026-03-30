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
		INSERT INTO music_plays (track_id, played_at, play_count)
		VALUES ($1, $2, $3)
	`, play.TrackID, play.PlayedAt, play.PlayCount)
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
		INSERT INTO unmatched_music_plays (music_brainz_id, title, played_at, play_count)
		VALUES ($1, $2, $3, $4)
	`, play.MusicBrainzID, play.Title, play.PlayedAt, play.PlayCount)
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

func GetTopArtists(ctx context.Context, limit int) ([]TopArtist, error) {
	metrics.LogQuery(ctx)

	query := `
		SELECT a.name,
		       (SELECT COUNT(*) FROM music_tracks t JOIN music_albums al ON al.id = t.album_id WHERE al.artist_id = a.id) AS track_count,
		       (SELECT COUNT(*) FROM music_albums al WHERE al.artist_id = a.id) AS album_count,
		       (SELECT SUM(max_pc) FROM (
		           SELECT MAX(p2.play_count) AS max_pc
		           FROM music_plays p2
		           JOIN music_tracks t2 ON t2.id = p2.track_id
		           JOIN music_albums al2 ON al2.id = t2.album_id
		           WHERE al2.artist_id = a.id
		           GROUP BY t2.id
		       ) sub) AS play_count,
		       MIN(p.played_at)::date AS first_played,
		       mr.path AS image_path
		FROM music_artists a
		JOIN music_albums al ON al.artist_id = a.id
		JOIN music_tracks t ON t.album_id = al.id
		JOIN music_plays p ON p.track_id = t.id
		LEFT JOIN media_relations mr ON mr.entity_type = 'artist' AND mr.entity_id = a.id AND mr.role = 'image'
		GROUP BY a.id, a.name, mr.path
		ORDER BY play_count DESC, a.sort_name`

	var args []any
	if limit > 0 {
		query += "\n\t\tLIMIT $1"
		args = append(args, limit)
	}

	var artists []TopArtist
	err := db.SelectContext(ctx, &artists, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get top artists: %w", err)
	}
	return artists, nil
}

func DeleteUnmatchedMusicPlay(ctx context.Context, id int) error {
	metrics.LogQuery(ctx)

	_, err := db.ExecContext(ctx, `DELETE FROM unmatched_music_plays WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete unmatched music play: %w", err)
	}
	return nil
}
