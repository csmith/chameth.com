package db

import (
	"context"
	"fmt"
	"time"
)

func GetVideoGameReviewByID(ctx context.Context, id int) (*VideoGameReview, error) {
	var review VideoGameReview
	err := db.GetContext(ctx, &review, "SELECT id, video_game_id, played_date, rating, playtime, completion_status, notes, published FROM video_game_reviews WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func GetVideoGameReviewsByVideoGameID(ctx context.Context, gameID int) ([]VideoGameReview, error) {
	var reviews []VideoGameReview
	err := db.SelectContext(ctx, &reviews, "SELECT id, video_game_id, played_date, rating, playtime, completion_status, notes, published FROM video_game_reviews WHERE video_game_id = $1 ORDER BY played_date DESC", gameID)
	if err != nil {
		return nil, err
	}
	return reviews, nil
}

func CreateVideoGameReview(ctx context.Context, gameID int, rating int, playedDate time.Time, playtime *int, completionStatus *string, published bool, notes string) (int, error) {
	var id int
	err := db.QueryRowContext(ctx, `
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
	_, err := db.ExecContext(ctx, `
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

	var result VideoGameReviewWithGameAndPoster
	err := db.GetContext(ctx, &result, query, reviewID)
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

	var results []VideoGameReviewWithGameAndPoster
	err := db.SelectContext(ctx, &results, query)
	if err != nil {
		return nil, err
	}

	return results, nil
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

	var results []VideoGameReviewWithGameAndPoster
	err := db.SelectContext(ctx, &results, query, limit)
	if err != nil {
		return nil, err
	}

	return results, nil
}

type VideoGameReviewWithGameAndPoster struct {
	VideoGameReview `db:"videogamereview"`
	VideoGame       `db:"videogame"`
	Poster          MediaRelationWithDetails `db:"poster"`
}
