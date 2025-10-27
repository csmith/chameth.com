package handlers

import (
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"path/filepath"

	"github.com/csmith/chameth.com/cmd/serve/assets"
	"github.com/csmith/chameth.com/cmd/serve/db"
)

func StaticAsset(w http.ResponseWriter, r *http.Request) {
	stat, err := fs.Stat(assets.Static, filepath.Join("static", r.URL.Path))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			NotFound(w, r)
			return
		}

		slog.Error("Failed to open static asset", "error", err)
		ServerError(w, r)
		return
	}

	if stat.IsDir() {
		// No directory listing!
		NotFound(w, r)
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=86400")
	http.ServeFileFS(w, r, assets.Static, filepath.Join("static", r.URL.Path))
}

func Stylesheet(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Path
	if path.Base(p) != assets.GetStylesheetPath() {
		w.Header().Set("Cache-Control", "private, no-cache, must-revalidate")
		http.Redirect(w, r, path.Join(path.Dir(p), assets.GetStylesheetPath()), http.StatusFound)
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(assets.GetStylesheet()))
}

func Media(w http.ResponseWriter, r *http.Request) {
	m, err := db.GetMediaBySlug(r.URL.Path)
	if err != nil {
		slog.Error("Failed to find media by slug", "error", err, "path", r.URL.Path)
		ServerError(w, r)
		return
	}

	w.Header().Set("Content-Type", m.ContentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(m.Data)
}
