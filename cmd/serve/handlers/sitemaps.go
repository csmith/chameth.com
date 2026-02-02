package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"chameth.com/chameth.com/cmd/serve/assets"
	"chameth.com/chameth.com/cmd/serve/content"
	"chameth.com/chameth.com/cmd/serve/db"
	"chameth.com/chameth.com/cmd/serve/templates"
)

func buildSiteMapData(ctx context.Context, pageData templates.PageData) (templates.SiteMapData, error) {
	poems, err := db.GetAllPoems(ctx)
	if err != nil {
		return templates.SiteMapData{}, fmt.Errorf("failed to get all poems: %w", err)
	}

	var poemDetails []templates.ContentDetails
	for _, p := range poems {
		poemDetails = append(poemDetails, templates.ContentDetails{
			Title: p.Title,
			Path:  p.Path,
			Date: templates.ContentDate{
				Iso:      p.Date.Format("2006-01-02"),
				Friendly: p.Date.Format("Jan 2, 2006"),
			},
		})
	}

	snippets, err := db.GetAllSnippets(ctx)
	if err != nil {
		return templates.SiteMapData{}, fmt.Errorf("failed to get all snippets: %w", err)
	}

	var snippetDetails []templates.SnippetDetails
	for _, s := range snippets {
		snippetDetails = append(snippetDetails, templates.SnippetDetails{
			Path: s.Path,
			Name: fmt.Sprintf("%s ➔ %s", s.Topic, s.Title),
		})
	}

	posts, err := db.GetAllPosts(ctx)
	if err != nil {
		return templates.SiteMapData{}, fmt.Errorf("failed to get all posts: %w", err)
	}

	var postDetails []templates.ContentDetails
	for _, p := range posts {
		postDetails = append(postDetails, templates.ContentDetails{
			Title: p.Title,
			Path:  p.Path,
			Date: templates.ContentDate{
				Iso:      p.Date.Format("2006-01-02"),
				Friendly: p.Date.Format("Jan 2, 2006"),
			},
		})
	}

	filmReviews, err := db.GetAllPublishedFilmReviewsWithFilmAndPosters(ctx)
	if err != nil {
		return templates.SiteMapData{}, fmt.Errorf("failed to get all film reviews: %w", err)
	}

	var filmDetails []templates.ContentDetails
	for _, review := range filmReviews {
		filmDetails = append(filmDetails, templates.ContentDetails{
			Title: review.Film.Title,
			Path:  review.Film.Path,
			Date: templates.ContentDate{
				Iso:      review.FilmReview.WatchedDate.Format("2006-01-02"),
				Friendly: review.FilmReview.WatchedDate.Format("Jan 2, 2006"),
			},
		})
	}

	filmLists, err := db.GetAllFilmLists(ctx)
	if err != nil {
		return templates.SiteMapData{}, fmt.Errorf("failed to get all film lists: %w", err)
	}

	var filmListDetails []templates.ContentDetails
	for _, list := range filmLists {
		filmListDetails = append(filmListDetails, templates.ContentDetails{
			Title: list.Title,
			Path:  list.Path,
		})
	}

	return templates.SiteMapData{
		Posts:     postDetails,
		Poems:     poemDetails,
		Snippets:  snippetDetails,
		Films:     filmDetails,
		FilmLists: filmListDetails,
		PageData:  pageData,
	}, nil
}

func HtmlSiteMap(w http.ResponseWriter, r *http.Request) {
	pageData := templates.PageData{
		Title:        "Sitemap · Chameth.com",
		Stylesheet:   assets.GetStylesheetPath(),
		CanonicalUrl: "https://chameth.com/sitemap/",
		RecentPosts:  content.RecentPosts(),
	}

	siteMapData, err := buildSiteMapData(r.Context(), pageData)
	if err != nil {
		slog.Error("Failed to build site map data", "error", err)
		ServerError(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderHtmlSiteMap(w, siteMapData)
	if err != nil {
		slog.Error("Failed to render site map template", "error", err)
	}
}

func XmlSiteMap(w http.ResponseWriter, r *http.Request) {
	siteMapData, err := buildSiteMapData(r.Context(), templates.PageData{})
	if err != nil {
		slog.Error("Failed to build site map data", "error", err)
		ServerError(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderXmlSiteMap(w, siteMapData)
	if err != nil {
		slog.Error("Failed to render site map template", "error", err)
	}
}
