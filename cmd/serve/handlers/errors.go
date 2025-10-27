package handlers

import (
	"log/slog"
	"net/http"

	"github.com/csmith/chameth.com/cmd/serve/assets"
	"github.com/csmith/chameth.com/cmd/serve/content"
	"github.com/csmith/chameth.com/cmd/serve/templates"
)

func NotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	err := templates.RenderNotFound(w, templates.NotFoundData{
		PageData: templates.PageData{
			Title:       "Not found · Chameth.com",
			Stylesheet:  assets.GetStylesheetPath(),
			RecentPosts: content.RecentPosts(),
		},
	})
	if err != nil {
		slog.Error("Failed to render not found template", "error", err)
	}
}

func ServerError(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	err := templates.RenderServerError(w, templates.ServerErrorData{
		PageData: templates.PageData{
			Title:       "Server error · Chameth.com",
			Stylesheet:  assets.GetStylesheetPath(),
			RecentPosts: content.RecentPosts(),
		},
	})
	if err != nil {
		slog.Error("Failed to render not found template", "error", err)
	}
}
