package contact

import (
	"chameth.com/chameth.com/features/routing"
	"go.temporal.io/sdk/client"
)

func RegisterRoutes(rm *routing.Manager, tc client.Client) {
	rm.Public.HandleFunc("POST /api/contact", handleJSON(tc))
	rm.Public.HandleFunc("POST /api/form/contact", handleForm(tc))
}
