package main

import (
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/csmith/chameth.com/cmd/serve/assets"
)

func handleStaticAsset(w http.ResponseWriter, r *http.Request) {
	stat, err := fs.Stat(assets.Static, filepath.Join("static", r.URL.Path))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			slog.Warn("Asset not found, falling through to 11ty", "path", r.URL.Path)
			http.FileServer(http.Dir(*files)).ServeHTTP(w, r)
			return
		}

		slog.Error("Failed to open static asset", "error", err)
		handleServerError(w, r)
		return
	}

	if stat.IsDir() {
		// No directory listing!
		handleNotFound(w, r)
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=86400")
	http.ServeFileFS(w, r, assets.Static, filepath.Join("static", r.URL.Path))
}
