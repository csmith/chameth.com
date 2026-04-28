package topalbums

import (
	"context"
	"fmt"

	"chameth.com/chameth.com/db"
)

type topAlbum struct {
	Name       string  `db:"name"`
	ArtistName string  `db:"artist_name"`
	TrackCount int     `db:"track_count"`
	PlayCount  int     `db:"play_count"`
	ImagePath  *string `db:"image_path"`
}

func query(ctx context.Context, limit int) ([]topAlbum, error) {
	query := `
		SELECT al.name,
		       ar.name AS artist_name,
		       COUNT(DISTINCT t.id) AS track_count,
		       SUM(max_play.max_pc) AS play_count,
		       mr.path AS image_path
		FROM music_albums al
		JOIN music_artists ar ON ar.id = al.artist_id
		JOIN music_tracks t ON t.album_id = al.id
		JOIN LATERAL (
		    SELECT MAX(p.play_count) AS max_pc
		    FROM music_plays p
		    WHERE p.track_id = t.id
		) max_play ON max_play.max_pc IS NOT NULL
		LEFT JOIN media_relations mr ON mr.entity_type = 'album' AND mr.entity_id = al.id AND mr.role = 'image'
		GROUP BY al.id, al.name, ar.name, mr.path
		ORDER BY play_count DESC, al.sort_name`

	var args []any
	if limit > 0 {
		query += "\n\t\tLIMIT $1"
		args = append(args, limit)
	}

	albums, err := db.Select[topAlbum](ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get top albums: %w", err)
	}
	return albums, nil
}
