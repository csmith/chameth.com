package videogames

import (
	"time"

	"chameth.com/chameth.com/features/media"
)

type VideoGameMetadata struct {
	ID        int    `db:"id"`
	Title     string `db:"title"`
	Published bool   `db:"published"`
	Path      string `db:"path"`
}

type VideoGame struct {
	VideoGameMetadata
	Platform string `db:"platform"`
	Overview string `db:"overview"`
}

type VideoGameReview struct {
	ID               int       `db:"id"`
	VideoGameID      int       `db:"video_game_id"`
	PlayedDate       time.Time `db:"played_date"`
	Rating           int       `db:"rating"`
	Playtime         *int      `db:"playtime"`
	CompletionStatus *string   `db:"completion_status"`
	Notes            string    `db:"notes"`
	Published        bool      `db:"published"`
}

type VideoGameWithReview struct {
	VideoGame
	Review *VideoGameReview
}

type VideoGameReviewWithGameAndPoster struct {
	VideoGameReview `db:"videogamereview"`
	VideoGame       `db:"videogame"`
	Poster          media.MediaRelationWithDetails `db:"poster"`
}
