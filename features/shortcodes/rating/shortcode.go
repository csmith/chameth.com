package rating

import (
	"embed"

	"chameth.com/chameth.com/assets"
	"chameth.com/chameth.com/features/shortcodes"
)

//go:embed star-empty.png star-flat.png star-half.png star.png
var starImages embed.FS

func RegisterShortcodes(mgr *shortcodes.Manager) {
	mgr.Register("rating", RenderFromText)
}

func RegisterAssets(mgr *assets.Manager) {
	mgr.AddStatic(starImages, "")
}
