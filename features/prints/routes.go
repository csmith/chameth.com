package prints

import (
	"chameth.com/chameth.com/features/routing"
)

func RegisterRoutes(rm *routing.Manager) {
	rm.Public.HandleFunc("GET /prints/{$}", handleList)
}
