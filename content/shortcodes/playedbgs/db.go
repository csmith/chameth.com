package playedbgs

import (
	"context"
	"time"

	"chameth.com/chameth.com/db"
)

func query(ctx context.Context, startDate, endDate time.Time) ([]db.BoardgameGameWithPlayCount, error) {
	return db.Select[db.BoardgameGameWithPlayCount](ctx, `
		SELECT
			g.name,
			g.year,
			mr.path AS image_path,
			COUNT(p.id) AS play_count
		FROM boardgame_games g
		JOIN boardgame_plays p ON p.game_id = g.id
		LEFT JOIN media_relations mr ON mr.entity_type = 'boardgame' AND mr.entity_id = g.bgg_id AND mr.role = 'image'
		WHERE p.date >= $1 AND p.date <= $2
		GROUP BY g.id, mr.path
		ORDER BY play_count DESC, g.name ASC
	`, startDate, endDate)
}
