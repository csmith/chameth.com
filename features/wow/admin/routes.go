package admin

import (
	"chameth.com/chameth.com/features/routing"
)

func RegisterRoutes(rm *routing.Manager) {
	rm.Admin.HandleFunc("GET /wow", ListCharactersHandler())
	rm.Admin.HandleFunc("POST /wow/import", ImportCharacterHandler())
	rm.Admin.HandleFunc("GET /wow/refresh/{id}", RefreshCharacterHandler())
}
