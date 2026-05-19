package admin

import (
	"chameth.com/chameth.com/features/routing"
)

func RegisterRoutes(rm *routing.Manager) {
	rm.Admin.HandleFunc("GET /poems", ListPoemsHandler())
	rm.Admin.HandleFunc("POST /poems", CreatePoemHandler())
	rm.Admin.HandleFunc("GET /poems/edit/{id}", EditPoemHandler())
	rm.Admin.HandleFunc("POST /poems/edit/{id}", UpdatePoemHandler())
}
