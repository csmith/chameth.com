package boardgames

import "time"

type BoardgameGame struct {
	ID          string `db:"id"`
	BggID       int    `db:"bgg_id"`
	Name        string `db:"name"`
	Year        int    `db:"year"`
	Status      string `db:"status"`
	IsExpansion bool   `db:"is_expansion"`
}

type BoardgameGameWithStats struct {
	Name       string     `db:"name"`
	Year       int        `db:"year"`
	ImagePath  *string    `db:"image_path"`
	PlayCount  int        `db:"play_count"`
	LastPlayed *time.Time `db:"last_played"`
}

type BoardgameGameWithPlayCount struct {
	Name      string  `db:"name"`
	Year      int     `db:"year"`
	ImagePath *string `db:"image_path"`
	PlayCount int     `db:"play_count"`
}

type BoardgamePlay struct {
	ID     string    `db:"id"`
	GameID string    `db:"game_id"`
	Date   time.Time `db:"date"`
}
