package features

import (
	"embed"

	"chameth.com/chameth.com/assets"
)

//go:embed */*/*.css
var assetsFS embed.FS

func init() {
	assets.Register(assetsFS, "features")
}
