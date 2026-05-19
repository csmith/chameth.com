package handlers

import (
	"chameth.com/chameth.com/assets"
	"chameth.com/chameth.com/features/routing"
)

func RegisterRoutes(rm *routing.Manager, assetsMgr *assets.Manager) {
	rm.Public.Handle("/", contentHandler(assets.StaticAssetHandler(assetsMgr)))
}
