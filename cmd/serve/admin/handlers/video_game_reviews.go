package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"chameth.com/chameth.com/cmd/serve/admin/templates"
	"chameth.com/chameth.com/cmd/serve/db"
)

func EditVideoGameReviewHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid review ID", http.StatusBadRequest)
			return
		}

		review, err := db.GetVideoGameReviewByID(id)
		if err != nil {
			http.Error(w, "Review not found", http.StatusNotFound)
			return
		}

		game, err := db.GetVideoGameByID(review.VideoGameID)
		if err != nil {
			http.Error(w, "Video game not found", http.StatusNotFound)
			return
		}

		playtime := ""
		if review.Playtime != nil {
			playtime = fmt.Sprintf("%d", *review.Playtime)
		}

		completionStatus := ""
		if review.CompletionStatus != nil {
			completionStatus = *review.CompletionStatus
		}

		data := templates.EditVideoGameReviewData{
			ID:               review.ID,
			VideoGameID:      game.ID,
			VideoGameTitle:   game.Title,
			PlayedDate:       review.PlayedDate.Format("2006-01-02"),
			Rating:           fmt.Sprintf("%d", review.Rating),
			Playtime:         playtime,
			CompletionStatus: completionStatus,
			Notes:            review.Notes,
			Published:        review.Published,
		}

		if err := templates.RenderEditVideoGameReview(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func CreateVideoGameReviewHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid video game ID", http.StatusBadRequest)
			return
		}

		_, err = db.GetVideoGameByID(id)
		if err != nil {
			http.Error(w, "Video game not found", http.StatusNotFound)
			return
		}

		reviewID, err := db.CreateVideoGameReview(id, 0, time.Now(), nil, nil, false, "")
		if err != nil {
			slog.Error("Failed to create video game review", "error", err)
			http.Error(w, "Failed to create video game review", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/video-game-reviews/edit/%d", reviewID), http.StatusSeeOther)
	}
}

func UpdateVideoGameReviewHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid review ID", http.StatusBadRequest)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		playedDate := r.FormValue("played_date")
		ratingStr := r.FormValue("rating")
		rating, err := strconv.Atoi(ratingStr)
		if err != nil {
			http.Error(w, "Invalid rating", http.StatusBadRequest)
			return
		}

		var playtime *int
		playtimeStr := r.FormValue("playtime")
		if playtimeStr != "" {
			pt, err := strconv.Atoi(playtimeStr)
			if err != nil {
				http.Error(w, "Invalid playtime", http.StatusBadRequest)
				return
			}
			playtime = &pt
		}

		var completionStatus *string
		cs := r.FormValue("completion_status")
		if cs != "" {
			completionStatus = &cs
		}

		notes := r.FormValue("notes")
		published := r.FormValue("published") == "true"

		if err := db.UpdateVideoGameReview(id, rating, playedDate, playtime, completionStatus, published, notes); err != nil {
			http.Error(w, "Failed to update video game review", http.StatusInternalServerError)
			return
		}

		review, err := db.GetVideoGameReviewByID(id)
		if err != nil {
			http.Error(w, "Failed to retrieve review", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/videogames/edit/%d", review.VideoGameID), http.StatusSeeOther)
	}
}
