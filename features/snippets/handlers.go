package snippets

import (
	"log/slog"
	"net/http"

	"chameth.com/chameth.com/content"
	"chameth.com/chameth.com/features/snippets/templates"
	parenttemplates "chameth.com/chameth.com/templates"
)

func SnippetHandler(w http.ResponseWriter, r *http.Request) {
	snippet, err := GetSnippetByPath(r.Context(), r.URL.Path)
	if err != nil {
		slog.Error("Failed to find snippet by path", "error", err, "path", r.URL.Path)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if snippet.Path != r.URL.Path {
		http.Redirect(w, r, snippet.Path, http.StatusPermanentRedirect)
		return
	}

	renderedContent, err := content.RenderContent(r.Context(), "snippet", 0, snippet.Content, snippet.Path)
	if err != nil {
		slog.Error("Failed to render markdown for snippet content", "snippet", snippet.Title, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderSnippet(w, templates.SnippetData{
		SnippetTitle:   snippet.Title,
		SnippetGroup:   snippet.Topic,
		SnippetContent: renderedContent,
		PageData:       content.CreatePageData(r.Context(), snippet.Title, snippet.Path, parenttemplates.OpenGraphHeaders{}),
	})
	if err != nil {
		slog.Error("Failed to render snippet template", "error", err, "path", r.URL.Path)
	}
}

func HandleList(w http.ResponseWriter, r *http.Request) {
	allSnippets, err := GetAllSnippets(r.Context())
	if err != nil {
		slog.Error("Failed to get all snippets", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var groups []templates.SnippetGroup
	for _, snippet := range allSnippets {
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
		PageData:      content.CreatePageData(r.Context(), "Snippets", "/snippets/", parenttemplates.OpenGraphHeaders{}),
	})
	if err != nil {
		slog.Error("Failed to render snippets template", "error", err)
	}
}
