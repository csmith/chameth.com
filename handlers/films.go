package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"math"
	"net/http"
	"strings"

	"chameth.com/chameth.com/content"
	"chameth.com/chameth.com/content/markdown"
	"chameth.com/chameth.com/content/shortcodes/filmlist"
	"chameth.com/chameth.com/content/shortcodes/rating"
	"chameth.com/chameth.com/content/shortcodes/syndication"
	"chameth.com/chameth.com/db"
	"chameth.com/chameth.com/templates"
)

func Film(w http.ResponseWriter, r *http.Request) {
	film, err := db.GetFilmWithPosterByPath(r.Context(), r.URL.Path)
	if err != nil {
		slog.Error("Failed to find film by path", "error", err, "path", r.URL.Path)
		ServerError(w, r)
		return
	}

	if film.Path != r.URL.Path {
		http.Redirect(w, r, film.Path, http.StatusPermanentRedirect)
		return
	}

	reviews, err := db.GetFilmReviewsByFilmID(r.Context(), film.ID)
	if err != nil {
		slog.Error("Failed to get film reviews", "film_id", film.ID, "error", err)
		ServerError(w, r)
		return
	}

	var publishedReviews []db.FilmReview
	for _, review := range reviews {
		if review.Published {
			publishedReviews = append(publishedReviews, review)
		}
	}

	var reviewData []templates.FilmReviewData
	for _, review := range publishedReviews {
		ratingHTML, err := rating.Render(review.Rating)
		if err != nil {
			slog.Error("Failed to render rating", "review_id", review.ID, "error", err)
			ServerError(w, r)
			return
		}

		reviewTextHTML, err := markdown.Render(review.ReviewText)
		if err != nil {
			slog.Error("Failed to render review text", "review_id", review.ID, "error", err)
			ServerError(w, r)
			return
		}

		reviewData = append(reviewData, templates.FilmReviewData{
			WatchedDate: review.WatchedDate.Format("2006-01-02"),
			Rating:      review.Rating,
			RatingHTML:  template.HTML(ratingHTML),
			IsRewatch:   review.IsRewatch,
			HasSpoilers: review.HasSpoilers,
			Content:     reviewTextHTML,
		})
	}

	renderedOverview, err := markdown.Render(film.Overview)
	if err != nil {
		slog.Error("Failed to render film overview", "film_id", film.ID, "error", err)
		ServerError(w, r)
		return
	}

	year := ""
	if film.Year != nil {
		year = fmt.Sprintf("%d", *film.Year)
	}

	timesWatched := len(publishedReviews)
	var ratingHTML string
	if len(publishedReviews) > 0 {
		var sum int
		for _, review := range publishedReviews {
			sum += review.Rating
		}
		averageRating := float64(sum) / float64(len(publishedReviews))
		roundedRating := int(math.Round(averageRating))
		stars, err := rating.Render(roundedRating)
		if err != nil {
			slog.Error("Failed to render rating", "error", err, "rating", roundedRating)
		} else {
			ratingHTML = stars
		}
	}

	posterPath := ""
	if film.PosterPath != nil {
		posterPath = *film.PosterPath
	}

	lists, err := db.GetFilmListsContainingFilm(r.Context(), film.ID)
	if err != nil {
		slog.Error("Failed to get film lists containing film", "film_id", film.ID, "error", err)
	}

	var filmLists []template.HTML
	for _, list := range lists {
		listHTML, err := filmlist.Render(r.Context(), list.ID)
		if err != nil {
			slog.Error("Failed to render film list", "list_id", list.ID, "error", err)
		} else {
			filmLists = append(filmLists, template.HTML(listHTML))
		}
	}

	syndicationHTML, err := syndication.Render(r.Context(), film.Path)
	if err != nil {
		slog.Error("Failed to render syndications", "path", film.Path, "error", err)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderFilm(w, templates.FilmData{
		Title:         film.Title,
		Year:          year,
		TMDBID:        film.TMDBID,
		Overview:      renderedOverview,
		Reviews:       reviewData,
		TimesWatched:  timesWatched,
		AverageRating: template.HTML(ratingHTML),
		PosterPath:    posterPath,
		FilmLists:     filmLists,
		Syndications:  template.HTML(syndicationHTML),
		PageData:      content.CreatePageData(fmt.Sprintf("%s (%s)", film.Title, year), film.Path, templates.OpenGraphHeaders{}),
	})
	if err != nil {
		slog.Error("Failed to render film template", "error", err, "path", r.URL.Path)
	}
}

func FilmList(w http.ResponseWriter, r *http.Request) {
	filmList, err := db.GetFilmListByPath(r.Context(), r.URL.Path)
	if err != nil {
		slog.Error("Failed to find film list by path", "error", err, "path", r.URL.Path)
		ServerError(w, r)
		return
	}

	if filmList.Path != r.URL.Path {
		http.Redirect(w, r, filmList.Path, http.StatusPermanentRedirect)
		return
	}

	entries, err := db.GetFilmListEntriesWithDetails(r.Context(), filmList.ID)
	if err != nil {
		slog.Error("Failed to get film list entries", "list_id", filmList.ID, "error", err)
		ServerError(w, r)
		return
	}

	renderedDescription, err := markdown.Render(filmList.Description)
	if err != nil {
		slog.Error("Failed to render film list description", "list_id", filmList.ID, "error", err)
		ServerError(w, r)
		return
	}

	filmListItems := make([]templates.FilmListItem, len(entries))
	for i, entry := range entries {
		year := ""
		if entry.Film.Year != nil {
			year = fmt.Sprintf("%d", *entry.Film.Year)
		}

		var ratingHTML string
		var ratingText string
		var lastWatched string
		if entry.AverageRating != nil {
			roundedRating := int(math.Round(*entry.AverageRating))
			ratingText = fmt.Sprintf("%d/10", roundedRating)
			stars, err := rating.Render(roundedRating)
			if err != nil {
				slog.Error("Failed to render rating", "error", err, "rating", roundedRating)
			} else {
				ratingHTML = stars
			}
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
			RatingHTML:   template.HTML(ratingHTML),
			LastWatched:  lastWatched,
		}
	}

	syndicationHTML, err := syndication.Render(r.Context(), filmList.Path)
	if err != nil {
		slog.Error("Failed to render syndications", "path", filmList.Path, "error", err)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderFilmList(w, templates.FilmListData{
		ListTitle:    filmList.Title,
		Description:  renderedDescription,
		Entries:      filmListItems,
		Syndications: template.HTML(syndicationHTML),
		PageData:     content.CreatePageData(filmList.Title, filmList.Path, templates.OpenGraphHeaders{}),
	})
	if err != nil {
		slog.Error("Failed to render film list template", "error", err, "path", r.URL.Path)
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

func SearchFilms(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(query) < 2 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "query must be at least 2 characters"}`))
		return
	}

	results, err := db.SearchFilms(r.Context(), query)
	if err != nil {
		slog.Error("Failed to search films", "query", query, "error", err)
		ServerError(w, r)
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
