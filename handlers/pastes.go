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

func Paste(w http.ResponseWriter, r *http.Request) {
	paste, err := db.GetPasteByPath(r.Context(), r.URL.Path)
	if err != nil {
		slog.Error("Failed to find paste by path", "error", err, "path", r.URL.Path)
		ServerError(w, r)
		return
	}

	if paste.Path != r.URL.Path {
		http.Redirect(w, r, paste.Path, http.StatusPermanentRedirect)
		return
	}

	// Check for raw query parameter
	if r.URL.Query().Has("raw") {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(paste.Content))
		return
	}

	// Construct markdown fenced code block
	var md strings.Builder
	md.WriteString("```")
	if paste.Language != "" {
		md.WriteString(paste.Language)
	}
	md.WriteString("\n")
	md.WriteString(paste.Content)
	if !strings.HasSuffix(paste.Content, "\n") {
		md.WriteString("\n")
	}
	md.WriteString("```")

	renderedContent, err := markdown.Render(md.String())
	if err != nil {
		slog.Error("Failed to render markdown for paste", "paste", paste.Title, "error", err)
		ServerError(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderPaste(w, templates.PasteData{
		Content:  renderedContent,
		Language: paste.Language,
		Size:     len(paste.Content),
		ArticleData: templates.ArticleData{
			ArticleTitle:   paste.Title,
			ArticleSummary: paste.Content,
			ArticleDate: templates.ArticleDate{
				Iso:         paste.Date.Format("2006-01-02"),
				Friendly:    paste.Date.Format("Jan 2, 2006"),
				ShowWarning: false,
			},
			PageData: content.CreatePageData(paste.Title, paste.Path, templates.OpenGraphHeaders{}),
		},
	})
	if err != nil {
		slog.Error("Failed to render paste template", "error", err, "path", r.URL.Path)
	}
}
