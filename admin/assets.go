package admin

import (
	adminAssets "chameth.com/chameth.com/admin/assets"
	"chameth.com/chameth.com/assets"
)

func RegisterAssets(mgr *assets.Manager) {
	mgr.AddAdminStatic(adminAssets.FS, "/assets")
}
