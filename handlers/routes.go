package handlers

import (
	"chameth.com/chameth.com/assets"
	"chameth.com/chameth.com/features/routing"
)

func RegisterRoutes(rm *routing.Manager, assetsMgr *assets.Manager) {
	rm.Public.Handle("GET /assets/stylesheets/", stylesheet(assetsMgr))
	rm.Public.Handle("GET /assets/scripts/", scripts(assetsMgr))
	rm.Public.Handle("/", contentHandler(StaticAsset(assetsMgr)))
}
