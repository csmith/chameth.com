package admin

import (
	"chameth.com/chameth.com/features/routing"
)

func RegisterRoutes(rm *routing.Manager) {
	rm.Admin.HandleFunc("GET /pages", ListPagesHandler())
	rm.Admin.HandleFunc("POST /pages", CreatePageHandler())
	rm.Admin.HandleFunc("GET /pages/edit/{id}", EditPageHandler())
	rm.Admin.HandleFunc("POST /pages/edit/{id}", UpdatePageHandler())
}
