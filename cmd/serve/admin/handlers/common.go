package handlers

import (
	"fmt"
	"net/http"

	"github.com/csmith/chameth.com/cmd/serve/admin/assets"
	"github.com/csmith/chameth.com/cmd/serve/admin/templates"
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
