package handlers

import (
	"net/http"

	"chameth.com/chameth.com/assets"
)

func RegisterRoutes(mux *http.ServeMux, assetsMgr *assets.Manager) {
	mux.Handle("GET /assets/stylesheets/", stylesheet(assetsMgr))
	mux.Handle("GET /assets/scripts/", scripts(assetsMgr))
	mux.Handle("/", contentHandler(StaticAsset(assetsMgr)))
}
