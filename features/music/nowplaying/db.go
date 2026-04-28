package nowplaying

import (
	"context"
	"fmt"

	"chameth.com/chameth.com/db"
)

func query(ctx context.Context) (*db.NowPlaying, error) {
	np, err := db.Get[db.NowPlaying](ctx, `
		SELECT ar.name AS artist_name,
		       t.name AS track_name,
		       al.name AS album_name,
		       mr.path AS image_path,
		       p.played_at
		FROM music_plays p
		JOIN music_tracks t ON t.id = p.track_id
		JOIN music_albums al ON al.id = t.album_id
		JOIN music_artists ar ON ar.id = al.artist_id
		LEFT JOIN media_relations mr ON mr.entity_type = 'album' AND mr.entity_id = al.id AND mr.role = 'image'
		ORDER BY p.played_at DESC
		LIMIT 1
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get now playing: %w", err)
	}
	return &np, nil
}
