package features

import (
	"embed"

	"chameth.com/chameth.com/assets"
)

//go:embed */*.css */*/*.css */*/*.js
var assetsFS embed.FS

func RegisterAssets(mgr *assets.Manager) {
	mgr.Add(assetsFS, "features")
}
