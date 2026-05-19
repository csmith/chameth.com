package assets

import (
	"net/http"
	"path"

	"chameth.com/chameth.com/features/routing"
)

func RegisterRoutes(rm *routing.Manager, mgr *Manager) {
	rm.Public.Handle("GET /assets/stylesheets/", stylesheetHandler(mgr))
	rm.Public.Handle("GET /assets/scripts/", scriptsHandler(mgr))
}

func StaticAssetHandler(mgr *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fsys, fsPath, ok := mgr.StaticAsset(r.URL.Path)
		if !ok {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Cache-Control", "public, max-age=86400")
		http.ServeFileFS(w, r, fsys, fsPath)
	}
}

func stylesheetHandler(mgr *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, checksum := mgr.Bundle(PublicCSS)
		stylesheetPath := checksum + ".css"

		p := r.URL.Path
		if path.Base(p) != stylesheetPath {
			w.Header().Set("Cache-Control", "private, no-cache, must-revalidate")
			http.Redirect(w, r, path.Join(path.Dir(p), stylesheetPath), http.StatusFound)
			return
		}

		content, _ := mgr.Bundle(PublicCSS)
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)
	}
}

func scriptsHandler(mgr *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, checksum := mgr.Bundle(PublicJS)
		scriptPath := checksum + ".js"

		p := r.URL.Path
		if path.Base(p) != scriptPath {
			w.Header().Set("Cache-Control", "private, no-cache, must-revalidate")
			http.Redirect(w, r, path.Join(path.Dir(p), scriptPath), http.StatusFound)
			return
		}

		content, _ := mgr.Bundle(PublicJS)
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)
	}
}
