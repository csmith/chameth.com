package playedalbums

import (
	"context"
	"fmt"
	"time"

	"chameth.com/chameth.com/db"
)

// query returns the limit albums with the most plays in [start, end), ranked,
// with each album's rank in the equivalent previous period [prevStart, start).
// previous_position is null when the album had no plays in that period.
//
// music_plays rows carry the track's lifetime play counter at the time of the
// row, not a per-play count, so a track's plays within a range are the
// difference between its maximum counters at the range boundaries. Counters
// are differenced per track, then summed per album: a track's counter is only
// comparable with itself.
func query(ctx context.Context, start, end, prevStart time.Time, limit int) ([]playedAlbum, error) {
	albums, err := db.Select[playedAlbum](ctx, `
		WITH track_counters AS (
		    SELECT p.track_id,
		           BOOL_OR(p.played_at >= $1) AS played_in_range,
		           COALESCE(MAX(p.play_count) FILTER (WHERE p.played_at < $2), 0)
		           - COALESCE(MAX(p.play_count) FILTER (WHERE p.played_at < $1), 0) AS track_plays,
		           COALESCE(MAX(p.play_count) FILTER (WHERE p.played_at < $1), 0)
		           - COALESCE(MAX(p.play_count) FILTER (WHERE p.played_at < $3), 0) AS previous_track_plays
		    FROM music_plays p
		    WHERE p.played_at < $2
		    GROUP BY p.track_id
		),
		current_plays AS (
		    SELECT t.album_id,
		           COUNT(*) FILTER (WHERE tc.played_in_range) AS track_count,
		           SUM(tc.track_plays) AS play_count
		    FROM track_counters tc
		    JOIN music_tracks t ON t.id = tc.track_id
		    GROUP BY t.album_id
		),
		previous_plays AS (
		    SELECT t.album_id,
		           SUM(tc.previous_track_plays) AS play_count
		    FROM track_counters tc
		    JOIN music_tracks t ON t.id = tc.track_id
		    GROUP BY t.album_id
		),
		current_ranked AS (
		    SELECT cp.album_id,
		           cp.track_count,
		           cp.play_count,
		           ROW_NUMBER() OVER (ORDER BY cp.play_count DESC, al.sort_name) AS position
		    FROM current_plays cp
		    JOIN music_albums al ON al.id = cp.album_id
		    WHERE cp.play_count > 0
		),
		previous_ranked AS (
		    SELECT pp.album_id,
		           ROW_NUMBER() OVER (ORDER BY pp.play_count DESC, al.sort_name) AS position
		    FROM previous_plays pp
		    JOIN music_albums al ON al.id = pp.album_id
		    WHERE pp.play_count > 0
		)
		SELECT cr.position,
		       al.name,
		       ar.name AS artist_name,
		       cr.track_count,
		       cr.play_count,
		       pr.position AS previous_position,
		       mr.path AS image_path
		FROM current_ranked cr
		JOIN music_albums al ON al.id = cr.album_id
		JOIN music_artists ar ON ar.id = al.artist_id
		LEFT JOIN previous_ranked pr ON pr.album_id = cr.album_id
		LEFT JOIN media_relations mr ON mr.entity_type = 'album' AND mr.entity_id = al.id AND mr.role = 'image'
		ORDER BY cr.position
		LIMIT $4
	`, start, end, prevStart, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get played albums: %w", err)
	}
	return albums, nil
}
