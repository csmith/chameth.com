package features

import (
	"embed"

	"chameth.com/chameth.com/assets"
)

//go:embed */*.css */*/*.css
var assetsFS embed.FS

func RegisterAssets(mgr *assets.Manager) {
	mgr.Add(assetsFS, "features")
}
