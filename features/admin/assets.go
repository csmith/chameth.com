package admin

import (
	"embed"

	"chameth.com/chameth.com/assets"
)

//go:embed assets/harper/*.*
var staticFS embed.FS

func RegisterAssets(mgr *assets.Manager) {
	mgr.AddAdminStatic(staticFS, "/assets")
}
