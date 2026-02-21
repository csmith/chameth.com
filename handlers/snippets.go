package handlers

import (
	"log/slog"
	"net/http"

	"chameth.com/chameth.com/content"
	"chameth.com/chameth.com/db"
	"chameth.com/chameth.com/templates"
)

func Snippet(w http.ResponseWriter, r *http.Request) {
	snippet, err := db.GetSnippetByPath(r.Context(), r.URL.Path)
	if err != nil {
		slog.Error("Failed to find snippet by path", "error", err, "path", r.URL.Path)
		ServerError(w, r)
		return
	}

	if snippet.Path != r.URL.Path {
		http.Redirect(w, r, snippet.Path, http.StatusPermanentRedirect)
		return
	}

	renderedContent, err := content.RenderContent(r.Context(), "snippet", 0, snippet.Content, snippet.Path)
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
		PageData:       content.CreatePageData(snippet.Title, snippet.Path, templates.OpenGraphHeaders{}),
	})
	if err != nil {
		slog.Error("Failed to render snippet template", "error", err, "path", r.URL.Path)
	}
}

func SnippetsList(w http.ResponseWriter, r *http.Request) {
	snippets, err := db.GetAllSnippets(r.Context())
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
			Path: snippet.Path,
		})
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderSnippets(w, templates.SnippetsData{
		SnippetGroups: groups,
		PageData:      content.CreatePageData("Snippets", "/snippets/", templates.OpenGraphHeaders{}),
	})
}
