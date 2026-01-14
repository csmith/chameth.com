package db

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"
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

func GetVideoGameByID(id int) (*VideoGame, error) {
	var game VideoGame
	err := db.Get(&game, "SELECT id, title, platform, overview, published, path FROM video_games WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &game, nil
}

func GetAllVideoGames() ([]VideoGame, error) {
	var games []VideoGame
	err := db.Select(&games, "SELECT id, title, platform, overview, published, path FROM video_games ORDER BY title")
	if err != nil {
		return nil, err
	}
	return games, nil
}

func GetAllVideoGamesWithReviews() ([]VideoGameWithReview, error) {
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

	rows, err := db.Query(query)
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

func CreateVideoGame(title, platform, overview, path string) (int, error) {
	var id int
	err := db.QueryRow(`
		INSERT INTO video_games (title, platform, overview, published, path)
		VALUES ($1, $2, $3, false, $4)
		RETURNING id
	`, title, platform, overview, path).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create video game: %w", err)
	}
	return id, nil
}

func UpdateVideoGame(id int, title, platform, overview, path string, published bool) error {
	_, err := db.Exec(`
		UPDATE video_games
		SET title = $1, platform = $2, overview = $3, published = $4, path = $5
		WHERE id = $6
	`, title, platform, overview, published, path, id)
	if err != nil {
		return fmt.Errorf("failed to update video game: %w", err)
	}
	return nil
}

func GetVideoGameByPath(path string) (*VideoGame, error) {
	var game VideoGame
	err := db.Get(&game, "SELECT id, title, platform, overview, published, path FROM video_games WHERE path = $1 OR path = $2", path, path+"/")
	if err != nil {
		return nil, err
	}
	return &game, nil
}

func DeleteVideoGame(id int) error {
	_, err := db.Exec("DELETE FROM video_games WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete video game: %w", err)
	}
	return nil
}
