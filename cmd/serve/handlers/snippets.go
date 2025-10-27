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

func Snippet(w http.ResponseWriter, r *http.Request) {
	snippet, err := db.GetSnippetBySlug(r.URL.Path)
	if err != nil {
		slog.Error("Failed to find snippet by slug", "error", err, "path", r.URL.Path)
		ServerError(w, r)
		return
	}

	if snippet.Slug != r.URL.Path {
		http.Redirect(w, r, snippet.Slug, http.StatusPermanentRedirect)
		return
	}

	renderedContent, err := content.RenderContent("snippet", 0, snippet.Content)
	if err != nil {
		slog.Error("Failed to render markdown for snippet content", "snippet", snippet.Title, "error", err)
		ServerError(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderSnippet(w, templates.SnippetData{
		SnippetTitle:   snippet.Title,
		SnippetGroup:   snippet.Topic,
		SnippetContent: renderedContent,
		PageData: templates.PageData{
			Title:        fmt.Sprintf("%s · Chameth.com", snippet.Title),
			Stylesheet:   assets.GetStylesheetPath(),
			CanonicalUrl: fmt.Sprintf("https://chameth.com%s", snippet.Slug),
			RecentPosts:  content.RecentPosts(),
		},
	})
	if err != nil {
		slog.Error("Failed to render snippet template", "error", err, "path", r.URL.Path)
	}
}

func SnippetsList(w http.ResponseWriter, r *http.Request) {
	snippets, err := db.GetAllSnippets()
	if err != nil {
		slog.Error("Failed to get all snippets", "error", err)
		ServerError(w, r)
		return
	}

	var groups []templates.SnippetGroup
	for _, snippet := range snippets {
		if len(groups) == 0 || groups[len(groups)-1].Name != snippet.Topic {
			groups = append(groups, templates.SnippetGroup{Name: snippet.Topic})
		}
		groups[len(groups)-1].Snippets = append(groups[len(groups)-1].Snippets, templates.SnippetDetails{
			Name: snippet.Title,
			Slug: snippet.Slug,
		})
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderSnippets(w, templates.SnippetsData{
		SnippetGroups: groups,
		PageData: templates.PageData{
			Title:        "Snippets · Chameth.com",
			Stylesheet:   assets.GetStylesheetPath(),
			CanonicalUrl: "https://chameth.com/snippets/",
			RecentPosts:  content.RecentPosts(),
		},
	})
}
