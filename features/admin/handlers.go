package admin

import (
	"net/http"

	"chameth.com/chameth.com/assets"
	"chameth.com/chameth.com/features/admin/templates"
	"chameth.com/chameth.com/features/routing"
)

func RegisterRoutes(rm *routing.Manager, assetsMgr *assets.Manager) {
	rm.Admin.HandleFunc("GET /{$}", handleIndex())
	rm.Admin.Handle("GET /assets/stylesheets/admin.css", handleAdminStylesheet(assetsMgr))
	rm.Admin.Handle("GET /assets/scripts/admin.js", handleAdminScript(assetsMgr))
	rm.Admin.HandleFunc("GET /", handleAssets(assetsMgr))
}

func handleIndex() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		data := templates.IndexData{}
		if err := templates.RenderIndex(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func handleAssets(mgr *assets.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fsys, fsPath, ok := mgr.StaticAssetWithFallback(r.URL.Path)
		if !ok {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Cache-Control", "public, max-age=86400")
		http.ServeFileFS(w, r, fsys, fsPath)
	}
}

func handleAdminStylesheet(mgr *assets.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		content, _ := mgr.Bundle(assets.AdminCSS)
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)
	}
}

func handleAdminScript(mgr *assets.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		content, _ := mgr.Bundle(assets.AdminJS)
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)
	}
}
