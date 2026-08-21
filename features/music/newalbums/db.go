package newalbums

import (
	"context"
	"fmt"
	"time"

	"chameth.com/chameth.com/db"
)

// query returns the albums first played in [start, end): each has at least one
// play in the range, and no track of it was ever played before start.
//
// music_plays rows carry the track's lifetime play counter at the time of the
// row, not a per-play count, so a track's plays before the range are normally
// the maximum counter logged before start. When a track has no rows before
// start, the counter on its first in-range row minus the rows logged in range
// recovers plays from before play logging began.
func query(ctx context.Context, start, end time.Time) ([]newAlbum, error) {
	albums, err := db.Select[newAlbum](ctx, `
		WITH track_stats AS (
		    SELECT p.track_id,
		           BOOL_OR(p.played_at >= $1) AS played_in_range,
		           COALESCE(MAX(p.play_count) FILTER (WHERE p.played_at < $2), 0)
		           - COALESCE(MAX(p.play_count) FILTER (WHERE p.played_at < $1), 0) AS track_plays,
		           GREATEST(
		               COALESCE(MAX(p.play_count) FILTER (WHERE p.played_at < $1), 0),
		               COALESCE(MAX(p.play_count) FILTER (WHERE p.played_at >= $1 AND p.played_at < $2), 0)
		               - COUNT(*) FILTER (WHERE p.played_at >= $1 AND p.played_at < $2)
		           ) AS plays_before_start
		    FROM music_plays p
		    WHERE p.played_at < $2
		    GROUP BY p.track_id
		)
		SELECT al.name,
		       ar.name AS artist_name,
		       SUM(ts.track_plays) AS play_count,
		       mr.path AS image_path
		FROM track_stats ts
		JOIN music_tracks t ON t.id = ts.track_id
		JOIN music_albums al ON al.id = t.album_id
		JOIN music_artists ar ON ar.id = al.artist_id
		LEFT JOIN media_relations mr ON mr.entity_type = 'album' AND mr.entity_id = al.id AND mr.role = 'image'
		GROUP BY al.id, ar.name, mr.path
		HAVING BOOL_OR(ts.played_in_range) AND MAX(ts.plays_before_start) = 0
		ORDER BY play_count DESC, al.sort_name
	`, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get new albums: %w", err)
	}
	return albums, nil
}
