package list

import (
	"context"

	"chameth.com/chameth.com/db"
	"chameth.com/chameth.com/features/boardgames"
)

func query(ctx context.Context) ([]boardgames.BoardgameGameWithStats, error) {
	return db.Select[boardgames.BoardgameGameWithStats](ctx, `
		SELECT
			g.name,
			g.year,
			mr.path AS image_path,
			COUNT(p.id) AS play_count,
			MAX(p.date) AS last_played
		FROM boardgame_games g
		JOIN boardgame_plays p ON p.game_id = g.id
		LEFT JOIN media_relations mr ON mr.entity_type = 'boardgame' AND mr.entity_id = g.bgg_id AND mr.role = 'image'
		WHERE g.is_expansion = false
		GROUP BY g.id, mr.path
		ORDER BY play_count DESC, g.name ASC
	`)
}
