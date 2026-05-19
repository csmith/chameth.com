package handlers

import (
	"net/http"

	"chameth.com/chameth.com/admin/assets"
	"chameth.com/chameth.com/admin/templates"
	publicAssets "chameth.com/chameth.com/assets"
	"chameth.com/chameth.com/features/routing"
)

func RegisterRoutes(rm *routing.Manager, assetsMgr *publicAssets.Manager) {
	rm.Admin.Handle("GET /assets/", http.StripPrefix("/assets/", AssetsHandler()))
	rm.Admin.HandleFunc("GET /{$}", IndexHandler())
	rm.Admin.HandleFunc("GET /", StaticAsset(assetsMgr))
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

func StaticAsset(mgr *publicAssets.Manager) http.HandlerFunc {
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
