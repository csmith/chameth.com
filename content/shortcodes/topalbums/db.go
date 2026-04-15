package topalbums

import (
	"context"
	"fmt"

	"chameth.com/chameth.com/db"
)

func query(ctx context.Context, limit int) ([]db.TopAlbum, error) {
	query := `
		SELECT al.name,
		       ar.name AS artist_name,
		       (SELECT COUNT(*) FROM music_tracks t WHERE t.album_id = al.id) AS track_count,
		       (SELECT SUM(max_pc) FROM (
		           SELECT MAX(p2.play_count) AS max_pc
		           FROM music_plays p2
		           JOIN music_tracks t2 ON t2.id = p2.track_id
		           WHERE t2.album_id = al.id
		           GROUP BY t2.id
		       ) sub) AS play_count,
		       mr.path AS image_path
		FROM music_albums al
		JOIN music_artists ar ON ar.id = al.artist_id
		JOIN music_tracks t ON t.album_id = al.id
		JOIN music_plays p ON p.track_id = t.id
		LEFT JOIN media_relations mr ON mr.entity_type = 'album' AND mr.entity_id = al.id AND mr.role = 'image'
		GROUP BY al.id, al.name, ar.name, mr.path
		ORDER BY play_count DESC, al.sort_name`

	var args []any
	if limit > 0 {
		query += "\n\t\tLIMIT $1"
		args = append(args, limit)
	}

	albums, err := db.Select[db.TopAlbum](ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get top albums: %w", err)
	}
	return albums, nil
}
