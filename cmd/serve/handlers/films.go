package handlers

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"

	"chameth.com/chameth.com/cmd/serve/assets"
	"chameth.com/chameth.com/cmd/serve/content"
	"chameth.com/chameth.com/cmd/serve/content/markdown"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes"
	"chameth.com/chameth.com/cmd/serve/db"
	"chameth.com/chameth.com/cmd/serve/templates"
)

func Film(w http.ResponseWriter, r *http.Request) {
	film, err := db.GetFilmByPath(r.URL.Path)
	if err != nil {
		slog.Error("Failed to find film by path", "error", err, "path", r.URL.Path)
		ServerError(w, r)
		return
	}

	if film.Path != r.URL.Path {
		http.Redirect(w, r, film.Path, http.StatusPermanentRedirect)
		return
	}

	reviews, err := db.GetFilmReviewsByFilmID(film.ID)
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

	var reviewHTMLs []template.HTML
	for _, review := range publishedReviews {
		reviewHTML, err := shortcodes.RenderFilmReview(review.ID)
		if err != nil {
			slog.Error("Failed to render film review", "review_id", review.ID, "error", err)
			ServerError(w, r)
			return
		}
		reviewHTMLs = append(reviewHTMLs, template.HTML(reviewHTML))
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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderFilm(w, templates.FilmData{
		Title:    film.Title,
		Year:     year,
		TMDBID:   film.TMDBID,
		Overview: renderedOverview,
		Reviews:  reviewHTMLs,
		PageData: templates.PageData{
			Title:        fmt.Sprintf("%s (%s) · Chameth.com", film.Title, year),
			Stylesheet:   assets.GetStylesheetPath(),
			CanonicalUrl: fmt.Sprintf("https://chameth.com%s", film.Path),
			RecentPosts:  content.RecentPosts(),
		},
	})
	if err != nil {
		slog.Error("Failed to render film template", "error", err, "path", r.URL.Path)
	}
}
