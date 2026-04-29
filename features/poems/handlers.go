package poems

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"chameth.com/chameth.com/content"
	"chameth.com/chameth.com/content/markdown"
	"chameth.com/chameth.com/features/poems/templates"
	parenttemplates "chameth.com/chameth.com/templates"
)

func PoemHandler(w http.ResponseWriter, r *http.Request) {
	poem, err := GetPoemByPath(r.Context(), r.URL.Path)
	if err != nil {
		slog.Error("Failed to find poem by path", "error", err, "path", r.URL.Path)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if poem.Path != r.URL.Path {
		http.Redirect(w, r, poem.Path, http.StatusPermanentRedirect)
		return
	}

	renderedComments, err := markdown.Render(poem.Notes)
	if err != nil {
		slog.Error("Failed to render markdown for poem comments", "poem", poem.Title, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderPoem(w, templates.PoemData{
		Poem:     strings.Split(poem.Poem, "\n"),
		Comments: renderedComments,
		ArticleData: parenttemplates.ArticleData{
			ArticleTitle:   poem.Title,
			ArticleSummary: poem.Poem,
			ArticleDate: parenttemplates.ArticleDate{
				Iso:         poem.Date.Format("2006-01-02"),
				Friendly:    poem.Date.Format("Jan 2, 2006"),
				ShowWarning: false,
			},
			EditLink: fmt.Sprintf("https://website-admin.yak-wall.ts.net/poems/edit/%d", poem.ID),
			PageData: content.CreatePageData(r.Context(), poem.Title, poem.Path, parenttemplates.OpenGraphHeaders{}),
		},
	})
	if err != nil {
		slog.Error("Failed to render poem template", "error", err, "path", r.URL.Path)
	}
}
