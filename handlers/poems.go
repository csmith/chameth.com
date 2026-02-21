package handlers

import (
	"log/slog"
	"net/http"
	"strings"

	"chameth.com/chameth.com/content"
	"chameth.com/chameth.com/content/markdown"
	"chameth.com/chameth.com/db"
	"chameth.com/chameth.com/templates"
)

func Poem(w http.ResponseWriter, r *http.Request) {
	poem, err := db.GetPoemByPath(r.Context(), r.URL.Path)
	if err != nil {
		slog.Error("Failed to find poem by path", "error", err, "path", r.URL.Path)
		ServerError(w, r)
		return
	}

	if poem.Path != r.URL.Path {
		http.Redirect(w, r, poem.Path, http.StatusPermanentRedirect)
		return
	}

	renderedComments, err := markdown.Render(poem.Notes)
	if err != nil {
		slog.Error("Failed to render markdown for poem comments", "poem", poem.Title, "error", err)
		ServerError(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderPoem(w, templates.PoemData{
		Poem:     strings.Split(poem.Poem, "\n"),
		Comments: renderedComments,
		ArticleData: templates.ArticleData{
			ArticleTitle:   poem.Title,
			ArticleSummary: poem.Poem,
			ArticleDate: templates.ArticleDate{
				Iso:         poem.Date.Format("2006-01-02"),
				Friendly:    poem.Date.Format("Jan 2, 2006"),
				ShowWarning: false,
			},
			PageData: content.CreatePageData(poem.Title, poem.Path, templates.OpenGraphHeaders{}),
		},
	})
	if err != nil {
		slog.Error("Failed to render poem template", "error", err, "path", r.URL.Path)
	}
}
