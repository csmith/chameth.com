package main

import (
	"net/http"
	"path"

	"github.com/csmith/chameth.com/cmd/serve/assets"
)

func serveStylesheet() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	})
}
