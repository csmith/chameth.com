package main

import (
	"log/slog"
	"net/http"

	"github.com/csmith/chameth.com/cmd/serve/templates"
)

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	err := templates.RenderNotFound(w, templates.NotFoundData{
		PageData: templates.PageData{
			Title:       "Not found · Chameth.com",
			Stylesheet:  compiledSheetPath,
			RecentPosts: recentPosts,
		},
	})
	if err != nil {
		slog.Error("Failed to render not found template", "error", err)
	}
}

func handlePGP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err := templates.RenderPGP(w, templates.PGPData{
		PageData: templates.PageData{
			Title:        "PGP information · Chameth.com",
			CanonicalUrl: "https://chameth.com/pgp/",
			Stylesheet:   compiledSheetPath,
			RecentPosts:  recentPosts,
		},
	})
	if err != nil {
		slog.Error("Failed to render pgp template", "error", err)
	}
}
