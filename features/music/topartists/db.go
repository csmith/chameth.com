package topartists

import (
	"context"
	"fmt"

	"chameth.com/chameth.com/db"
)

type topArtist struct {
	Name       string  `db:"name"`
	TrackCount int     `db:"track_count"`
	AlbumCount int     `db:"album_count"`
	PlayCount  int     `db:"play_count"`
	ImagePath  *string `db:"image_path"`
}

func query(ctx context.Context, limit int) ([]topArtist, error) {
	query := `
		SELECT a.name,
		       COUNT(DISTINCT t.id) AS track_count,
		       COUNT(DISTINCT al.id) AS album_count,
		       SUM(max_play.max_pc) AS play_count,
		       mr.path AS image_path
		FROM music_artists a
		JOIN music_albums al ON al.artist_id = a.id
		JOIN music_tracks t ON t.album_id = al.id
		JOIN LATERAL (
		    SELECT MAX(p.play_count) AS max_pc
		    FROM music_plays p
		    WHERE p.track_id = t.id
		) max_play ON max_play.max_pc IS NOT NULL
		LEFT JOIN media_relations mr ON mr.entity_type = 'artist' AND mr.entity_id = a.id AND mr.role = 'image'
		GROUP BY a.id, a.name, mr.path
		ORDER BY play_count DESC, a.sort_name`

	var args []any
	if limit > 0 {
		query += "\n\t\tLIMIT $1"
		args = append(args, limit)
	}

	artists, err := db.Select[topArtist](ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get top artists: %w", err)
	}
	return artists, nil
}
