package rating

import (
	"embed"

	"chameth.com/chameth.com/assets"
	"chameth.com/chameth.com/features/shortcodes"
)

//go:embed star-empty.png star-flat.png star-half.png star.png
var starImages embed.FS

func init() {
	shortcodes.Register("rating", RenderFromText)
}

func RegisterAssets(mgr *assets.Manager) {
	mgr.AddStatic(starImages, "")
}
