package videogames

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"chameth.com/chameth.com/db"
)

func generateVideoGamePath(title string) string {
	lowered := strings.ToLower(title)
	replaced := strings.Map(func(r rune) rune {
		if r == ' ' {
			return '-'
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return '-'
	}, lowered)
	cleaned := regexp.MustCompile(`-+`).ReplaceAllString(replaced, "-")
	cleaned = regexp.MustCompile(`^-+|-+$`).ReplaceAllString(cleaned, "")
	return "/videogames/" + cleaned + "/"
}

func GetVideoGameByID(ctx context.Context, id int) (*VideoGame, error) {
	game, err := db.Get[VideoGame](ctx, "SELECT id, title, platform, overview, published, path FROM video_games WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &game, nil
}

func GetAllVideoGames(ctx context.Context) ([]VideoGame, error) {
	return db.Select[VideoGame](ctx, "SELECT id, title, platform, overview, published, path FROM video_games ORDER BY title")
}

func GetAllVideoGamesWithReviews(ctx context.Context) ([]VideoGameWithReview, error) {
	query := `
		SELECT
			vg.id, vg.title, vg.platform, vg.overview, vg.published, vg.path,
			vgr.id as review_id, vgr.video_game_id as review_video_game_id, vgr.played_date, vgr.rating, vgr.playtime, vgr.completion_status, vgr.notes, vgr.published as review_published
		FROM video_games vg
		LEFT JOIN LATERAL (
			SELECT * FROM video_game_reviews
			WHERE video_game_id = vg.id
			ORDER BY played_date DESC
			LIMIT 1
		) vgr ON true
		ORDER BY vg.title
	`

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var games []VideoGameWithReview
	for rows.Next() {
		var vg VideoGame
		var review VideoGameReview
		var reviewID, reviewVideoGameID sql.NullInt64
		var playedDate sql.NullTime
		var rating sql.NullInt64
		var playtime sql.NullInt64
		var completionStatus sql.NullString
		var notes sql.NullString
		var reviewPublished sql.NullBool

		err := rows.Scan(
			&vg.ID, &vg.Title, &vg.Platform, &vg.Overview, &vg.Published, &vg.Path,
			&reviewID, &reviewVideoGameID, &playedDate, &rating, &playtime, &completionStatus, &notes, &reviewPublished,
		)
		if err != nil {
			return nil, err
		}

		if reviewID.Valid {
			review.ID = int(reviewID.Int64)
			review.VideoGameID = int(reviewVideoGameID.Int64)
			review.PlayedDate = playedDate.Time
			review.Rating = int(rating.Int64)
			if playtime.Valid {
				pt := int(playtime.Int64)
				review.Playtime = &pt
			}
			if completionStatus.Valid {
				review.CompletionStatus = &completionStatus.String
			}
			review.Notes = notes.String
			review.Published = reviewPublished.Bool
			games = append(games, VideoGameWithReview{VideoGame: vg, Review: &review})
		} else {
			games = append(games, VideoGameWithReview{VideoGame: vg, Review: nil})
		}
	}

	return games, nil
}

func CreateVideoGame(ctx context.Context, title, platform, overview, path string) (int, error) {
	var id int
	err := db.QueryRow(ctx, `
		INSERT INTO video_games (title, platform, overview, published, path)
		VALUES ($1, $2, $3, false, $4)
		RETURNING id
	`, title, platform, overview, path).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create video game: %w", err)
	}
	return id, nil
}

func UpdateVideoGame(ctx context.Context, id int, title, platform, overview, path string, published bool) error {
	_, err := db.Exec(ctx, `
		UPDATE video_games
		SET title = $1, platform = $2, overview = $3, published = $4, path = $5
		WHERE id = $6
	`, title, platform, overview, published, path, id)
	if err != nil {
		return fmt.Errorf("failed to update video game: %w", err)
	}
	return nil
}

func GetVideoGameByPath(ctx context.Context, path string) (*VideoGame, error) {
	game, err := db.Get[VideoGame](ctx, "SELECT id, title, platform, overview, published, path FROM video_games WHERE path = $1 OR path = $2", path, path+"/")
	if err != nil {
		return nil, err
	}
	return &game, nil
}

func DeleteVideoGame(ctx context.Context, id int) error {
	_, err := db.Exec(ctx, "DELETE FROM video_games WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete video game: %w", err)
	}
	return nil
}

func GetVideoGameReviewByID(ctx context.Context, id int) (*VideoGameReview, error) {
	review, err := db.Get[VideoGameReview](ctx, "SELECT id, video_game_id, played_date, rating, playtime, completion_status, notes, published FROM video_game_reviews WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func GetVideoGameReviewsByVideoGameID(ctx context.Context, gameID int) ([]VideoGameReview, error) {
	return db.Select[VideoGameReview](ctx, "SELECT id, video_game_id, played_date, rating, playtime, completion_status, notes, published FROM video_game_reviews WHERE video_game_id = $1 ORDER BY played_date DESC", gameID)
}

func CreateVideoGameReview(ctx context.Context, gameID int, rating int, playedDate time.Time, playtime *int, completionStatus *string, published bool, notes string) (int, error) {
	var id int
	err := db.QueryRow(ctx, `
		INSERT INTO video_game_reviews (video_game_id, rating, played_date, playtime, completion_status, notes, published)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, gameID, rating, playedDate, playtime, completionStatus, notes, published).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create video game review: %w", err)
	}
	return id, nil
}

func UpdateVideoGameReview(ctx context.Context, id int, rating int, playedDate string, playtime *int, completionStatus *string, published bool, notes string) error {
	_, err := db.Exec(ctx, `
		UPDATE video_game_reviews
		SET rating = $1, played_date = $2, playtime = $3, completion_status = $4, notes = $5, published = $6
		WHERE id = $7
	`, rating, playedDate, playtime, completionStatus, notes, published, id)
	if err != nil {
		return fmt.Errorf("failed to update video game review: %w", err)
	}
	return nil
}

func GetVideoGameReviewWithGameAndPoster(ctx context.Context, reviewID int) (*VideoGameReviewWithGameAndPoster, error) {
	query := `
		SELECT
			vgr.id as "videogamereview.id", vgr.video_game_id as "videogamereview.video_game_id", vgr.played_date as "videogamereview.played_date",
			vgr.rating as "videogamereview.rating", vgr.playtime as "videogamereview.playtime", vgr.completion_status as "videogamereview.completion_status",
			vgr.notes as "videogamereview.notes", vgr.published as "videogamereview.published",
			vg.id as "videogame.id", vg.title as "videogame.title", vg.platform as "videogame.platform",
			vg.overview as "videogame.overview", vg.published as "videogame.published",
			mr.path as "poster.path", mr.media_id as "poster.media_id", mr.description as "poster.description",
			mr.caption as "poster.caption", mr.role as "poster.role", mr.entity_type as "poster.entity_type",
			mr.entity_id as "poster.entity_id",
			m.id as "poster.id", m.content_type as "poster.content_type", m.original_filename as "poster.original_filename",
			m.width as "poster.width", m.height as "poster.height", m.parent_media_id as "poster.parent_media_id"
		FROM video_game_reviews vgr
		JOIN video_games vg ON vgr.video_game_id = vg.id
		JOIN media_relations mr ON mr.entity_type = 'videogame' AND mr.entity_id = vg.id AND mr.role = 'poster'
		JOIN media m ON mr.media_id = m.id
		WHERE vgr.id = $1
	`

	result, err := db.Get[VideoGameReviewWithGameAndPoster](ctx, query, reviewID)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func GetAllPublishedVideoGameReviewsWithGameAndPosters(ctx context.Context) ([]VideoGameReviewWithGameAndPoster, error) {
	query := `
		SELECT
			vgr.id as "videogamereview.id", vgr.video_game_id as "videogamereview.video_game_id", vgr.played_date as "videogamereview.played_date",
			vgr.rating as "videogamereview.rating", vgr.playtime as "videogamereview.playtime", vgr.completion_status as "videogamereview.completion_status",
			vgr.notes as "videogamereview.notes", vgr.published as "videogamereview.published",
			vg.id as "videogame.id", vg.title as "videogame.title", vg.platform as "videogame.platform",
			vg.overview as "videogame.overview", vg.published as "videogame.published", vg.path as "videogame.path",
			mr.path as "poster.path", mr.media_id as "poster.media_id", mr.description as "poster.description",
			mr.caption as "poster.caption", mr.role as "poster.role", mr.entity_type as "poster.entity_type",
			mr.entity_id as "poster.entity_id",
			m.id as "poster.id", m.content_type as "poster.content_type", m.original_filename as "poster.original_filename",
			m.width as "poster.width", m.height as "poster.height", m.parent_media_id as "poster.parent_media_id"
		FROM video_game_reviews vgr
		JOIN video_games vg ON vgr.video_game_id = vg.id
		JOIN media_relations mr ON mr.entity_type = 'videogame' AND mr.entity_id = vg.id AND mr.role = 'poster'
		JOIN media m ON mr.media_id = m.id
		WHERE vgr.published = true
		ORDER BY vgr.played_date DESC
	`

	return db.Select[VideoGameReviewWithGameAndPoster](ctx, query)
}

func GetRecentPublishedVideoGameReviewsWithGameAndPosters(ctx context.Context, limit int) ([]VideoGameReviewWithGameAndPoster, error) {
	query := `
		SELECT
			vgr.id as "videogamereview.id", vgr.video_game_id as "videogamereview.video_game_id", vgr.played_date as "videogamereview.played_date",
			vgr.rating as "videogamereview.rating", vgr.playtime as "videogamereview.playtime", vgr.completion_status as "videogamereview.completion_status",
			vgr.notes as "videogamereview.notes", vgr.published as "videogamereview.published",
			vg.id as "videogame.id", vg.title as "videogame.title", vg.platform as "videogame.platform",
			vg.overview as "videogame.overview", vg.published as "videogame.published", vg.path as "videogame.path",
			mr.path as "poster.path", mr.media_id as "poster.media_id", mr.description as "poster.description",
			mr.caption as "poster.caption", mr.role as "poster.role", mr.entity_type as "poster.entity_type",
			mr.entity_id as "poster.entity_id",
			m.id as "poster.id", m.content_type as "poster.content_type", m.original_filename as "poster.original_filename",
			m.width as "poster.width", m.height as "poster.height", m.parent_media_id as "poster.parent_media_id"
		FROM video_game_reviews vgr
		JOIN video_games vg ON vgr.video_game_id = vg.id
		LEFT JOIN media_relations mr ON mr.entity_type = 'videogame' AND mr.entity_id = vg.id AND mr.role = 'poster'
		LEFT JOIN media m ON mr.media_id = m.id
		WHERE vgr.published = true
		ORDER BY vgr.played_date DESC
		LIMIT $1
	`

	return db.Select[VideoGameReviewWithGameAndPoster](ctx, query, limit)
}
