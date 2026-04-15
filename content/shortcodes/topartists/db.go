package topartists

import (
	"context"
	"fmt"

	"chameth.com/chameth.com/db"
)

func query(ctx context.Context, limit int) ([]db.TopArtist, error) {
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

	artists, err := db.Select[db.TopArtist](ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get top artists: %w", err)
	}
	return artists, nil
}
