package contact

import (
	"chameth.com/chameth.com/features/routing"
)

func RegisterRoutes(rm *routing.Manager) {
	rm.Public.HandleFunc("POST /api/contact", handleJSON)
	rm.Public.HandleFunc("POST /api/form/contact", handleForm)
}
