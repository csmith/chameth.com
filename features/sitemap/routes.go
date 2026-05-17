package sitemap

import (
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /sitemap.xml", handleXml)
	mux.HandleFunc("GET /sitemap/{$}", handleHtml)
}
