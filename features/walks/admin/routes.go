package admin

import (
	"chameth.com/chameth.com/features/routing"
)

func RegisterRoutes(rm *routing.Manager) {
	rm.Admin.HandleFunc("POST /api/walks/import", ImportWalksHandler())
}
