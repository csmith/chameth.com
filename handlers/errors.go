package handlers

import (
	"log/slog"
	"net/http"

	"chameth.com/chameth.com/content"
	"chameth.com/chameth.com/templates"
)

func NotFound(w http.ResponseWriter, r *http.Request) {
	slog.Warn("Serving not found response", "url", r.URL.String(), "method", r.Method)
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	err := templates.RenderNotFound(w, templates.NotFoundData{
		PageData: content.CreatePageData(r.Context(), "Not found", "", templates.OpenGraphHeaders{}),
	})
	if err != nil {
		slog.Error("Failed to render not found template", "error", err)
	}
}

func ServerError(w http.ResponseWriter, r *http.Request) {
	slog.Warn("Serving server error response", "url", r.URL.String(), "method", r.Method)
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	err := templates.RenderServerError(w, templates.ServerErrorData{
		PageData: content.CreatePageData(r.Context(), "Server error", "", templates.OpenGraphHeaders{}),
	})
	if err != nil {
		slog.Error("Failed to render not found template", "error", err)
	}
}
