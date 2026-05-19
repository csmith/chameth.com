package admin

import (
	"chameth.com/chameth.com/features/routing"
)

func RegisterRoutes(rm *routing.Manager) {
	rm.Admin.HandleFunc("GET /goimports", ListGoImportsHandler())
	rm.Admin.HandleFunc("POST /goimports", CreateGoImportHandler())
	rm.Admin.HandleFunc("GET /goimports/edit/{id}", EditGoImportHandler())
	rm.Admin.HandleFunc("POST /goimports/edit/{id}", UpdateGoImportHandler())
}
