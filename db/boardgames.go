package db

import (
	"context"
	"fmt"
	"time"

	"chameth.com/chameth.com/metrics"
)

func UpsertBoardgameGame(ctx context.Context, game BoardgameGame) error {
	metrics.LogQuery(ctx)

	_, err := db.ExecContext(ctx, `
		INSERT INTO boardgame_games (id, bgg_id, name, year, status, is_expansion)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id)
		DO UPDATE SET
			bgg_id = EXCLUDED.bgg_id,
			name = EXCLUDED.name,
			year = EXCLUDED.year,
			status = EXCLUDED.status,
			is_expansion = EXCLUDED.is_expansion
	`, game.ID, game.BggID, game.Name, game.Year, game.Status, game.IsExpansion)

	if err != nil {
		return fmt.Errorf("failed to upsert boardgame game: %w", err)
	}

	return nil
}

func GetBoardgameGamesWithStats(ctx context.Context) ([]BoardgameGameWithStats, error) {
	metrics.LogQuery(ctx)
	var games []BoardgameGameWithStats
	err := db.SelectContext(ctx, &games, `
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
	if err != nil {
		return nil, err
	}
	return games, nil
}

func GetBoardgameGamesWithPlayCountByDateRange(ctx context.Context, startDate, endDate time.Time) ([]BoardgameGameWithPlayCount, error) {
	metrics.LogQuery(ctx)
	var games []BoardgameGameWithPlayCount
	err := db.SelectContext(ctx, &games, `
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
	if err != nil {
		return nil, err
	}
	return games, nil
}

func UpsertBoardgamePlay(ctx context.Context, play BoardgamePlay) error {
	metrics.LogQuery(ctx)

	_, err := db.ExecContext(ctx, `
		INSERT INTO boardgame_plays (id, game_id, date)
		VALUES ($1, $2, $3)
		ON CONFLICT (id)
		DO UPDATE SET
			game_id = EXCLUDED.game_id,
			date = EXCLUDED.date
	`, play.ID, play.GameID, play.Date)

	if err != nil {
		return fmt.Errorf("failed to upsert boardgame play: %w", err)
	}

	return nil
}
