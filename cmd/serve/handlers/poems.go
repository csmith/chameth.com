package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/csmith/chameth.com/cmd/serve/assets"
	"github.com/csmith/chameth.com/cmd/serve/content"
	"github.com/csmith/chameth.com/cmd/serve/db"
	"github.com/csmith/chameth.com/cmd/serve/templates"
)

func Poem(w http.ResponseWriter, r *http.Request) {
	poem, err := db.GetPoemByPath(r.URL.Path)
	if err != nil {
		slog.Error("Failed to find poem by path", "error", err, "path", r.URL.Path)
		ServerError(w, r)
		return
	}

	if poem.Path != r.URL.Path {
		http.Redirect(w, r, poem.Path, http.StatusPermanentRedirect)
		return
	}

	renderedComments, err := content.RenderMarkdown(poem.Notes)
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
			PageData: templates.PageData{
				Title:        fmt.Sprintf("%s · Chameth.com", poem.Title),
				Stylesheet:   assets.GetStylesheetPath(),
				CanonicalUrl: fmt.Sprintf("https://chameth.com%s", poem.Path),
				RecentPosts:  content.RecentPosts(),
			},
		},
	})
	if err != nil {
		slog.Error("Failed to render poem template", "error", err, "path", r.URL.Path)
	}
}
