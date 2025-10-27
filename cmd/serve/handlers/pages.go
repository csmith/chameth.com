package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/csmith/chameth.com/cmd/serve/assets"
	"github.com/csmith/chameth.com/cmd/serve/content"
	"github.com/csmith/chameth.com/cmd/serve/db"
	"github.com/csmith/chameth.com/cmd/serve/templates"
)

func StaticPage(w http.ResponseWriter, r *http.Request) {
	page, err := db.GetStaticPageBySlug(r.URL.Path)
	if err != nil {
		slog.Error("Failed to find static page by slug", "error", err, "path", r.URL.Path)
		ServerError(w, r)
		return
	}

	if page.Slug != r.URL.Path {
		http.Redirect(w, r, page.Slug, http.StatusPermanentRedirect)
		return
	}

	renderedContent, err := content.RenderContent("staticpage", page.ID, page.Content)
	if err != nil {
		slog.Error("Failed to render static page content", "page", page.Title, "error", err)
		ServerError(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderStaticPage(w, templates.StaticPageData{
		StaticTitle:   page.Title,
		StaticContent: renderedContent,
		PageData: templates.PageData{
			Title:        fmt.Sprintf("%s · Chameth.com", page.Title),
			Stylesheet:   assets.GetStylesheetPath(),
			CanonicalUrl: fmt.Sprintf("https://chameth.com%s", page.Slug),
			RecentPosts:  content.RecentPosts(),
		},
	})
	if err != nil {
		slog.Error("Failed to render static page template", "error", err, "path", r.URL.Path)
	}
}
