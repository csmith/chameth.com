package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"chameth.com/chameth.com/admin/templates"
	"chameth.com/chameth.com/db"
	"github.com/csmith/aca"
)

func ListVideoGamesHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		games, err := db.GetAllVideoGamesWithReviews(r.Context())
		if err != nil {
			http.Error(w, "Failed to retrieve video games", http.StatusInternalServerError)
			return
		}

		gameSummaries := make([]templates.VideoGameSummary, len(games))
		for i, game := range games {
			rating := ""
			if game.Review != nil {
				rating = fmt.Sprintf("%d/10", game.Review.Rating)
			}

			gameSummaries[i] = templates.VideoGameSummary{
				ID:        game.ID,
				Title:     game.Title,
				Platform:  game.Platform,
				Rating:    rating,
				Published: game.Published,
			}
		}

		data := templates.ListVideoGamesData{
			VideoGames: gameSummaries,
		}

		if err := templates.RenderListVideoGames(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func CreateVideoGameHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		gen, err := aca.NewDefaultGenerator()
		if err != nil {
			http.Error(w, "Failed to generate name", http.StatusInternalServerError)
			return
		}
		name := gen.Generate()
		path := fmt.Sprintf("/videogames/%s/", name)
		gameID, err := db.CreateVideoGame(r.Context(), name, "", "", path)
		if err != nil {
			slog.Error("Failed to create video game", "error", err)
			http.Error(w, "Failed to create video game", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/videogames/edit/%d", gameID), http.StatusSeeOther)
	}
}

func EditVideoGameHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid video game ID", http.StatusBadRequest)
			return
		}

		game, err := db.GetVideoGameByID(r.Context(), id)
		if err != nil {
			http.Error(w, "Video game not found", http.StatusNotFound)
			return
		}

		mediaRelations, err := db.GetMediaRelationsForEntity(r.Context(), "videogame", id)
		if err != nil {
			http.Error(w, "Failed to retrieve media relations", http.StatusInternalServerError)
			return
		}

		var poster *templates.MediaItem
		for _, rel := range mediaRelations {
			if rel.Role != nil && *rel.Role == "poster" {
				poster = &templates.MediaItem{
					ID:               rel.MediaID,
					OriginalFilename: rel.ContentType,
					Width:            rel.Width,
					Height:           rel.Height,
					ContentType:      rel.ContentType,
				}
				break
			}
		}

		reviews, err := db.GetVideoGameReviewsByVideoGameID(r.Context(), id)
		if err != nil {
			http.Error(w, "Failed to retrieve video game reviews", http.StatusInternalServerError)
			return
		}

		reviewSummaries := make([]templates.VideoGameReviewSummary, len(reviews))
		for i, review := range reviews {
			playtime := ""
			if review.Playtime != nil {
				playtime = fmt.Sprintf("%d", *review.Playtime)
			}

			completionStatus := ""
			if review.CompletionStatus != nil {
				completionStatus = *review.CompletionStatus
			}

			reviewSummaries[i] = templates.VideoGameReviewSummary{
				ID:               review.ID,
				PlayedDate:       review.PlayedDate.Format("2006-01-02"),
				Rating:           fmt.Sprintf("%d", review.Rating),
				Playtime:         playtime,
				CompletionStatus: completionStatus,
				Notes:            review.Notes,
				Published:        review.Published,
			}
		}

		data := templates.EditVideoGameData{
			ID:        game.ID,
			Title:     game.Title,
			Platform:  game.Platform,
			Overview:  game.Overview,
			Published: game.Published,
			Path:      game.Path,
			Poster:    poster,
			Reviews:   reviewSummaries,
		}

		if err := templates.RenderEditVideoGame(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func UpdateVideoGameHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid video game ID", http.StatusBadRequest)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		path := r.FormValue("path")
		title := r.FormValue("title")
		platform := r.FormValue("platform")
		overview := r.FormValue("overview")
		published := r.FormValue("published") == "true"

		if title == "" {
			http.Error(w, "Title is required", http.StatusBadRequest)
			return
		}

		if err := db.UpdateVideoGame(r.Context(), id, title, platform, overview, path, published); err != nil {
			http.Error(w, "Failed to update video game", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/videogames/edit/%d", id), http.StatusSeeOther)
	}
}

func DeleteVideoGameHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid video game ID", http.StatusBadRequest)
			return
		}

		if err := db.DeleteVideoGame(r.Context(), id); err != nil {
			slog.Error("Failed to delete video game", "error", err, "id", id)
			http.Error(w, "Failed to delete video game", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/videogames", http.StatusSeeOther)
	}
}
