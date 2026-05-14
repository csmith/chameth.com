package feeds

import (
	"embed"

	"chameth.com/chameth.com/assets"
)

//go:embed feeds.css feeds.xsl
var staticFS embed.FS

func RegisterAssets(mgr *assets.Manager) {
	mgr.AddStatic(staticFS, "")
}
