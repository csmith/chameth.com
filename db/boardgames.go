package db

import (
	"context"
	"fmt"

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
