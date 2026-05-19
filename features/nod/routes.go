package nod

import (
	"chameth.com/chameth.com/features/routing"
)

func RegisterRoutes(rm *routing.Manager) {
	rm.Public.HandleFunc("POST /api/nod", handleJSON)
	rm.Public.HandleFunc("POST /api/form/nod", handleForm)
}
