package sitemap

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"chameth.com/chameth.com/content"
	"chameth.com/chameth.com/features/films"
	"chameth.com/chameth.com/features/pages"
	"chameth.com/chameth.com/features/poems"
	"chameth.com/chameth.com/features/posts"
	"chameth.com/chameth.com/features/snippets"
	"chameth.com/chameth.com/templates"
)

func buildSiteMapData(ctx context.Context, pageData templates.PageData) (SiteMapData, error) {
	poemDetails, err := poems.SitemapEntries(ctx)
	if err != nil {
		return SiteMapData{}, err
	}

	snippetDetails, err := snippets.SitemapEntries(ctx)
	if err != nil {
		return SiteMapData{}, err
	}

	allPosts, err := posts.GetAllPosts(ctx)
	if err != nil {
		return SiteMapData{}, fmt.Errorf("failed to get all posts: %w", err)
	}

	var postDetails []templates.ContentDetails
	for _, p := range allPosts {
		postDetails = append(postDetails, templates.ContentDetails{
			Title: p.Title,
			Path:  p.Path,
			Date: templates.ContentDate{
				Iso:      p.Date.Format("2006-01-02"),
				Friendly: p.Date.Format("Jan 2, 2006"),
			},
		})
	}

	filmReviews, err := films.GetAllPublishedFilmReviewsWithFilmAndPosters(ctx)
	if err != nil {
		return SiteMapData{}, fmt.Errorf("failed to get all film reviews: %w", err)
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

	filmLists, err := films.GetAllFilmLists(ctx)
	if err != nil {
		return SiteMapData{}, fmt.Errorf("failed to get all film lists: %w", err)
	}

	var filmListDetails []templates.ContentDetails
	for _, list := range filmLists {
		filmListDetails = append(filmListDetails, templates.ContentDetails{
			Title: list.Title,
			Path:  list.Path,
		})
	}

	sitemapPages, err := pages.GetSitemapStaticPages(ctx)
	if err != nil {
		return SiteMapData{}, fmt.Errorf("failed to get sitemap pages: %w", err)
	}

	var pageDetails []SiteMapPageDetails
	for _, p := range sitemapPages {
		pageDetails = append(pageDetails, SiteMapPageDetails{
			Path:      p.Path,
			Frequency: *p.SitemapFrequency,
			Priority:  fmt.Sprintf("%.1f", *p.SitemapPriority),
		})
	}

	return SiteMapData{
		Posts:     postDetails,
		Poems:     poemDetails,
		Snippets:  snippetDetails,
		Films:     filmDetails,
		FilmLists: filmListDetails,
		Pages:     pageDetails,
		PageData:  pageData,
	}, nil
}

func HandleHtml(w http.ResponseWriter, r *http.Request) {
	siteMapData, err := buildSiteMapData(r.Context(), content.CreatePageData(r.Context(), "Sitemap", "/sitemap/", templates.OpenGraphHeaders{}))
	if err != nil {
		slog.Error("Failed to build site map data", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = renderHtmlSiteMap(w, siteMapData)
	if err != nil {
		slog.Error("Failed to render site map template", "error", err)
	}
}

func HandleXml(w http.ResponseWriter, r *http.Request) {
	siteMapData, err := buildSiteMapData(r.Context(), templates.PageData{})
	if err != nil {
		slog.Error("Failed to build site map data", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = renderXmlSiteMap(w, siteMapData)
	if err != nil {
		slog.Error("Failed to render site map template", "error", err)
	}
}
