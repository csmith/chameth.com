package pastes

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"chameth.com/chameth.com/content"
	"chameth.com/chameth.com/content/markdown"
	"chameth.com/chameth.com/features/pastes/templates"
	parenttemplates "chameth.com/chameth.com/templates"
)

func PasteHandler(w http.ResponseWriter, r *http.Request) {
	paste, err := GetPasteByPath(r.Context(), r.URL.Path)
	if err != nil {
		slog.Error("Failed to find paste by path", "error", err, "path", r.URL.Path)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if paste.Path != r.URL.Path {
		http.Redirect(w, r, paste.Path, http.StatusPermanentRedirect)
		return
	}

	if r.URL.Query().Has("raw") {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(paste.Content))
		return
	}

	delimiter := "```"
	for strings.Contains(paste.Content, delimiter) {
		delimiter += "`"
	}

	var md strings.Builder
	md.WriteString(delimiter)
	if paste.Language != "" {
		md.WriteString(paste.Language)
	}
	md.WriteString("\n")
	md.WriteString(paste.Content)
	if !strings.HasSuffix(paste.Content, "\n") {
		md.WriteString("\n")
	}
	md.WriteString(delimiter)

	renderedContent, err := markdown.Render(md.String())
	if err != nil {
		slog.Error("Failed to render markdown for paste", "paste", paste.Title, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderPaste(w, templates.PasteData{
		Content:  renderedContent,
		Language: paste.Language,
		Size:     len(paste.Content),
		ArticleData: parenttemplates.ArticleData{
			ArticleTitle:   paste.Title,
			ArticleSummary: paste.Content,
			ArticleDate: parenttemplates.ArticleDate{
				Iso:         paste.Date.Format("2006-01-02"),
				Friendly:    paste.Date.Format("Jan 2, 2006"),
				ShowWarning: false,
			},
			EditLink: fmt.Sprintf("https://website-admin.yak-wall.ts.net/pastes/edit/%d", paste.ID),
			PageData: content.CreatePageData(r.Context(), paste.Title, paste.Path, parenttemplates.OpenGraphHeaders{}),
		},
	})
	if err != nil {
		slog.Error("Failed to render paste template", "error", err, "path", r.URL.Path)
	}
}
