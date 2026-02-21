package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"chameth.com/chameth.com/admin/templates"
	"chameth.com/chameth.com/db"
)

func EditFilmReviewHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid review ID", http.StatusBadRequest)
			return
		}

		review, err := db.GetFilmReviewByID(r.Context(), id)
		if err != nil {
			http.Error(w, "Review not found", http.StatusNotFound)
			return
		}

		film, err := db.GetFilmByID(r.Context(), review.FilmID)
		if err != nil {
			http.Error(w, "Film not found", http.StatusNotFound)
			return
		}

		data := templates.EditFilmReviewData{
			ID:          review.ID,
			FilmID:      film.ID,
			FilmTitle:   film.Title,
			WatchedDate: review.WatchedDate.Format("2006-01-02"),
			Rating:      fmt.Sprintf("%d", review.Rating),
			IsRewatch:   review.IsRewatch,
			HasSpoilers: review.HasSpoilers,
			ReviewText:  review.ReviewText,
			Published:   review.Published,
		}

		if err := templates.RenderEditFilmReview(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func CreateFilmReviewHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid film ID", http.StatusBadRequest)
			return
		}

		_, err = db.GetFilmByID(r.Context(), id)
		if err != nil {
			http.Error(w, "Film not found", http.StatusNotFound)
			return
		}

		reviewID, err := db.CreateFilmReview(r.Context(), id, 0, time.Now(), false, false, false, "")
		if err != nil {
			slog.Error("Failed to create film review", "error", err)
			http.Error(w, "Failed to create film review", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/film-reviews/edit/%d", reviewID), http.StatusSeeOther)
	}
}

func UpdateFilmReviewHandler() func(http.ResponseWriter, *http.Request) {
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

		watchedDate := r.FormValue("watched_date")
		ratingStr := r.FormValue("rating")
		rating, err := strconv.Atoi(ratingStr)
		if err != nil {
			http.Error(w, "Invalid rating", http.StatusBadRequest)
			return
		}
		isRewatch := r.FormValue("is_rewatch") == "true"
		hasSpoilers := r.FormValue("has_spoilers") == "true"
		reviewText := r.FormValue("review_text")
		published := r.FormValue("published") == "true"

		if err := db.UpdateFilmReview(r.Context(), id, rating, watchedDate, isRewatch, hasSpoilers, published, reviewText); err != nil {
			http.Error(w, "Failed to update film review", http.StatusInternalServerError)
			return
		}

		review, err := db.GetFilmReviewByID(r.Context(), id)
		if err != nil {
			http.Error(w, "Failed to retrieve review", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/films/edit/%d", review.FilmID), http.StatusSeeOther)
	}
}
