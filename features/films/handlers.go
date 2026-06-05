package films

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strings"

	"chameth.com/chameth.com/content"
	"chameth.com/chameth.com/content/markdown"
	"chameth.com/chameth.com/features/films/templates"
	"chameth.com/chameth.com/features/shortcodes/rating"
	maintemplates "chameth.com/chameth.com/templates"
)

func FilmPage(w http.ResponseWriter, r *http.Request) {
	film, err := GetFilmWithPosterByPath(r.Context(), r.URL.Path)
	if err != nil {
		slog.Error("Failed to find film by path", "error", err, "path", r.URL.Path)
		http.Error(w, "Film not found", http.StatusInternalServerError)
		return
	}

	if film.Path != r.URL.Path {
		http.Redirect(w, r, film.Path, http.StatusPermanentRedirect)
		return
	}

	reviews, err := GetFilmReviewsByFilmID(r.Context(), film.ID)
	if err != nil {
		slog.Error("Failed to get film reviews", "film_id", film.ID, "error", err)
		http.Error(w, "Failed to get reviews", http.StatusInternalServerError)
		return
	}

	var publishedReviews []FilmReview
	for _, review := range reviews {
		if review.Published {
			publishedReviews = append(publishedReviews, review)
		}
	}

	var reviewData []templates.FilmReviewData
	for _, review := range publishedReviews {
		reviewTextHTML, err := markdown.Render(review.ReviewText)
		if err != nil {
			slog.Error("Failed to render review text", "review_id", review.ID, "error", err)
			http.Error(w, "Failed to render review", http.StatusInternalServerError)
			return
		}

		reviewData = append(reviewData, templates.FilmReviewData{
			WatchedDate: review.WatchedDate.Format("2006-01-02"),
			Rating:      review.Rating,
			IsRewatch:   review.IsRewatch,
			HasSpoilers: review.HasSpoilers,
			Content:     reviewTextHTML,
		})
	}

	renderedOverview, err := markdown.Render(film.Overview)
	if err != nil {
		slog.Error("Failed to render film overview", "film_id", film.ID, "error", err)
		http.Error(w, "Failed to render overview", http.StatusInternalServerError)
		return
	}

	year := ""
	if film.Year != nil {
		year = fmt.Sprintf("%d", *film.Year)
	}

	timesWatched := len(publishedReviews)
	var averageRating int
	if len(publishedReviews) > 0 {
		var sum int
		for _, review := range publishedReviews {
			sum += review.Rating
		}
		averageRating = int(math.Round(float64(sum) / float64(len(publishedReviews))))
	}

	posterPath := ""
	if film.PosterPath != nil {
		posterPath = *film.PosterPath
	}

	lists, err := GetFilmListsContainingFilm(r.Context(), film.ID)
	if err != nil {
		slog.Error("Failed to get film lists containing film", "film_id", film.ID, "error", err)
	}

	var filmListIDs []int
	for _, list := range lists {
		filmListIDs = append(filmListIDs, list.ID)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderFilm(w, templates.FilmData{
		FilmTitle:     film.Title,
		Year:          year,
		TMDBID:        film.TMDBID,
		Overview:      renderedOverview,
		Reviews:       reviewData,
		TimesWatched:  timesWatched,
		AverageRating: averageRating,
		PosterPath:    posterPath,
		FilmLists:     filmListIDs,
		PageData:      content.CreatePageData(r.Context(), fmt.Sprintf("%s (%s)", film.Title, year), film.Path, maintemplates.OpenGraphHeaders{}),
	})
	if err != nil {
		slog.Error("Failed to render film template", "error", err, "path", r.URL.Path)
	}
}

func FilmListPage(w http.ResponseWriter, r *http.Request) {
	filmList, err := GetFilmListByPath(r.Context(), r.URL.Path)
	if err != nil {
		slog.Error("Failed to find film list by path", "error", err, "path", r.URL.Path)
		http.Error(w, "Film list not found", http.StatusInternalServerError)
		return
	}

	if filmList.Path != r.URL.Path {
		http.Redirect(w, r, filmList.Path, http.StatusPermanentRedirect)
		return
	}

	entries, err := GetFilmListEntriesWithDetails(r.Context(), filmList.ID)
	if err != nil {
		slog.Error("Failed to get film list entries", "list_id", filmList.ID, "error", err)
		http.Error(w, "Failed to get entries", http.StatusInternalServerError)
		return
	}

	renderedDescription, err := markdown.Render(filmList.Description)
	if err != nil {
		slog.Error("Failed to render film list description", "list_id", filmList.ID, "error", err)
		http.Error(w, "Failed to render description", http.StatusInternalServerError)
		return
	}

	filmListItems := make([]templates.FilmListItem, len(entries))
	for i, entry := range entries {
		year := ""
		if entry.Film.Year != nil {
			year = fmt.Sprintf("%d", *entry.Film.Year)
		}

		var roundedRating int
		var ratingText string
		var lastWatched string
		if entry.AverageRating != nil {
			roundedRating = int(math.Round(*entry.AverageRating))
			ratingText = fmt.Sprintf("%d/10", roundedRating)
		}
		if entry.LastWatched != nil {
			lastWatched = entry.LastWatched.Format("January 2, 2006")
		}

		filmListItems[i] = templates.FilmListItem{
			Position:     entry.Position,
			PosterPath:   entry.Poster.Path,
			FilmPath:     entry.Film.Path,
			Title:        entry.Film.Title,
			Year:         year,
			TimesWatched: entry.TimesWatched,
			RatingText:   ratingText,
			Rating:       roundedRating,
			LastWatched:  lastWatched,
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderFilmList(w, templates.FilmListData{
		ListTitle:   filmList.Title,
		Description: renderedDescription,
		Entries:     filmListItems,
		PageData:    content.CreatePageData(r.Context(), filmList.Title, filmList.Path, maintemplates.OpenGraphHeaders{}),
	})
	if err != nil {
		slog.Error("Failed to render film list template", "error", err, "path", r.URL.Path)
	}
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(query) < 2 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "query must be at least 2 characters"}`))
		return
	}

	results, err := SearchFilms(r.Context(), query)
	if err != nil {
		slog.Error("Failed to search films", "query", query, "error", err)
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	response := make([]filmSearchResult, len(results))
	for i, result := range results {
		var ratingHTML string
		if result.AverageRating != nil {
			roundedRating := int(math.Round(*result.AverageRating))
			stars, err := rating.Render(roundedRating)
			if err != nil {
				slog.Error("Failed to render rating", "error", err, "rating", roundedRating)
			} else {
				ratingHTML = stars
			}
		}

		response[i] = filmSearchResult{
			ID:           result.ID,
			Title:        result.Title,
			Path:         result.Path,
			PosterPath:   result.PosterPath,
			TimesWatched: result.TimesWatched,
			RatingHTML:   ratingHTML,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		slog.Error("Failed to encode search results", "error", err)
	}
}

type filmSearchResult struct {
	ID           int     `json:"id"`
	Title        string  `json:"title"`
	Path         string  `json:"path"`
	PosterPath   *string `json:"poster_path"`
	TimesWatched int     `json:"times_watched"`
	RatingHTML   string  `json:"rating_html"`
}
