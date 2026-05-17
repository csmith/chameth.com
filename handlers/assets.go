package handlers

import (
	"log/slog"
	"net/http"
	"path"

	"chameth.com/chameth.com/assets"
	"chameth.com/chameth.com/features/media"
)

func StaticAsset(mgr *assets.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fsys, fsPath, ok := mgr.StaticAsset(r.URL.Path)
		if !ok {
			NotFound(w, r)
			return
		}

		w.Header().Set("Cache-Control", "public, max-age=86400")
		http.ServeFileFS(w, r, fsys, fsPath)
	}
}

func stylesheet(mgr *assets.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, checksum := mgr.Bundle(assets.PublicCSS)
		stylesheetPath := checksum + ".css"

		p := r.URL.Path
		if path.Base(p) != stylesheetPath {
			w.Header().Set("Cache-Control", "private, no-cache, must-revalidate")
			http.Redirect(w, r, path.Join(path.Dir(p), stylesheetPath), http.StatusFound)
			return
		}

		content, _ := mgr.Bundle(assets.PublicCSS)
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)
	}
}

func scripts(mgr *assets.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, checksum := mgr.Bundle(assets.PublicJS)
		scriptPath := checksum + ".js"

		p := r.URL.Path
		if path.Base(p) != scriptPath {
			w.Header().Set("Cache-Control", "private, no-cache, must-revalidate")
			http.Redirect(w, r, path.Join(path.Dir(p), scriptPath), http.StatusFound)
			return
		}

		content, _ := mgr.Bundle(assets.PublicJS)
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)
	}
}

func Media(w http.ResponseWriter, r *http.Request) {
	m, err := media.GetMediaByPath(r.Context(), r.URL.Path)
	if err != nil {
		slog.Error("Failed to find media by path", "error", err, "path", r.URL.Path)
		ServerError(w, r)
		return
	}

	w.Header().Set("Content-Type", m.ContentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(m.Data)
}
