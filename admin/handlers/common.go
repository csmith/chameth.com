package handlers

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"

	"chameth.com/chameth.com/admin/assets"
	"chameth.com/chameth.com/admin/templates"
	publicAssets "chameth.com/chameth.com/assets"
)

func RedirectHandler(hostname func() string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		httpsURL := fmt.Sprintf("https://%s%s", hostname(), r.URL.Path)
		if r.URL.RawQuery != "" {
			httpsURL += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
	}
}

func AssetsHandler() http.Handler {
	return http.FileServer(http.FS(assets.FS))
}

func IndexHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		data := templates.IndexData{}
		if err := templates.RenderIndex(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func StaticAsset(w http.ResponseWriter, r *http.Request) {
	stat, err := fs.Stat(publicAssets.Static, filepath.Join("static", r.URL.Path))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			http.NotFound(w, r)
			return
		}

		slog.Error("Failed to open static asset", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if stat.IsDir() {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=86400")
	http.ServeFileFS(w, r, publicAssets.Static, filepath.Join("static", r.URL.Path))
}
