package pages

import (
	"log/slog"
	"net/http"

	"chameth.com/chameth.com/content"
	pagetemplates "chameth.com/chameth.com/features/pages/templates"
	parenttemplates "chameth.com/chameth.com/templates"
)

func StaticPageHandler(w http.ResponseWriter, r *http.Request) {
	page, err := GetStaticPageByPath(r.Context(), r.URL.Path)
	if err != nil {
		slog.Error("Failed to find static page by path", "error", err, "path", r.URL.Path)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if page.Path != r.URL.Path {
		http.Redirect(w, r, page.Path, http.StatusPermanentRedirect)
		return
	}

	if page.Raw {
		renderedContent, err := content.RenderContent(r.Context(), "rawpage", page.ID, page.Content, page.Path)
		if err != nil {
			slog.Error("Failed to render raw page content", "page", page.Title, "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		err = pagetemplates.RenderRawPage(w, pagetemplates.RawPageData{
			RawContent: renderedContent,
			PageData:   content.CreatePageData(r.Context(), page.Title, page.Path, parenttemplates.OpenGraphHeaders{}),
		})
		if err != nil {
			slog.Error("Failed to render raw page template", "error", err, "path", r.URL.Path)
		}
		return
	}

	renderedContent, err := content.RenderContent(r.Context(), "staticpage", page.ID, page.Content, page.Path)
	if err != nil {
		slog.Error("Failed to render static page content", "page", page.Title, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	err = pagetemplates.RenderStaticPage(w, pagetemplates.StaticPageData{
		StaticTitle:   page.Title,
		StaticContent: renderedContent,
		PageData:      content.CreatePageData(r.Context(), page.Title, page.Path, parenttemplates.OpenGraphHeaders{}),
	})
	if err != nil {
		slog.Error("Failed to render static page template", "error", err, "path", r.URL.Path)
	}
}
