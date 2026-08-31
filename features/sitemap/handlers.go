package sitemap

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

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
		if strings.Count(p.Path, "/") > 2 {
			continue
		}
		pageDetails = append(pageDetails, SiteMapPageDetails{
			Title:     p.Title,
			Path:      p.Path,
			Frequency: *p.SitemapFrequency,
			Priority:  fmt.Sprintf("%.1f", *p.SitemapPriority),
		})
	}

	pageDetails = append(pageDetails,
		SiteMapPageDetails{Title: "Posts", Path: "/posts/", Frequency: "daily", Priority: "0.2"},
		SiteMapPageDetails{Title: "3D Prints", Path: "/prints/", Frequency: "monthly", Priority: "0.2"},
		SiteMapPageDetails{Title: "Projects", Path: "/projects/", Frequency: "monthly", Priority: "0.5"},
		SiteMapPageDetails{Title: "Sitemap", Path: "/sitemap/", Frequency: "daily", Priority: "0.2", CurrentPage: true},
		SiteMapPageDetails{Title: "Snippets", Path: "/snippets/", Frequency: "weekly", Priority: "0.2"},
	)

	slices.SortFunc(pageDetails, func(a, b SiteMapPageDetails) int {
		if a.Path < b.Path {
			return -1
		}
		if a.Path > b.Path {
			return 1
		}
		return 0
	})

	pageTree, err := buildPageTree(ctx)
	if err != nil {
		return SiteMapData{}, err
	}

	return SiteMapData{
		Posts:     postDetails,
		Poems:     poemDetails,
		Snippets:  snippetDetails,
		Films:     filmDetails,
		FilmLists: filmListDetails,
		Pages:     pageDetails,
		PageTree:  pageTree,
		PageData:  pageData,
	}, nil
}

// codeDefinedChildren maps the path of a database-defined page to hard-coded
// child pages for the HTML sitemap tree. The children are served by code
// routes rather than staticpages rows, so the nesting cannot be expressed in
// the database. Children of pages that do not exist are dropped.
var codeDefinedChildren = map[string][]SiteMapPageDetails{
	"/feeds": {{Title: "Post feed builder", Path: "/feeds/posts/build/"}},
}

// buildPageTree builds the hierarchy of published pages for the HTML sitemap,
// nesting any page with a parent underneath it. A top-level page is included if
// it opts into the sitemap (frequency + priority) or has children to show.
func buildPageTree(ctx context.Context) ([]*SiteMapPageDetails, error) {
	allPages, err := pages.GetAllStaticPages(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get pages for tree: %w", err)
	}

	detailsByID := make(map[int]*SiteMapPageDetails, len(allPages))
	for _, p := range allPages {
		detail := &SiteMapPageDetails{Title: p.Title, Path: p.Path}
		if p.SitemapFrequency != nil {
			detail.Frequency = *p.SitemapFrequency
		}
		if p.SitemapPriority != nil {
			detail.Priority = fmt.Sprintf("%.1f", *p.SitemapPriority)
		}
		detailsByID[p.ID] = detail
	}

	for _, p := range allPages {
		if p.ParentID == nil {
			continue
		}
		if parent, ok := detailsByID[*p.ParentID]; ok {
			parent.Children = append(parent.Children, detailsByID[p.ID])
		}
	}

	for _, p := range allPages {
		children, ok := codeDefinedChildren[strings.TrimSuffix(p.Path, "/")]
		if !ok {
			continue
		}
		parent := detailsByID[p.ID]
		for i := range children {
			parent.Children = append(parent.Children, &children[i])
		}
	}

	var tree []*SiteMapPageDetails
	for _, p := range allPages {
		if p.ParentID != nil {
			if _, ok := detailsByID[*p.ParentID]; ok {
				continue
			}
		}
		detail := detailsByID[p.ID]
		if (p.SitemapFrequency != nil && p.SitemapPriority != nil) || len(detail.Children) > 0 {
			tree = append(tree, detail)
		}
	}

	tree = append(tree,
		&SiteMapPageDetails{Title: "Posts", Path: "/posts/"},
		&SiteMapPageDetails{Title: "3D Prints", Path: "/prints/"},
		&SiteMapPageDetails{Title: "Projects", Path: "/projects/"},
		&SiteMapPageDetails{Title: "Sitemap", Path: "/sitemap/", CurrentPage: true},
		&SiteMapPageDetails{Title: "Snippets", Path: "/snippets/"},
	)

	sortPageTree(tree)
	return tree, nil
}

func sortPageTree(nodes []*SiteMapPageDetails) {
	slices.SortFunc(nodes, func(a, b *SiteMapPageDetails) int {
		return strings.Compare(a.Path, b.Path)
	})
	for _, node := range nodes {
		sortPageTree(node.Children)
	}
}

func handleHtml(w http.ResponseWriter, r *http.Request) {
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

func handleXml(w http.ResponseWriter, r *http.Request) {
	siteMapData, err := buildSiteMapData(r.Context(), templates.PageData{SiteURL: templates.SiteURL()})
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
