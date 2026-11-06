package nod

import (
	"chameth.com/chameth.com/features/routing"
	"go.temporal.io/sdk/client"
)

func RegisterRoutes(rm *routing.Manager, tc client.Client) {
	rm.Public.HandleFunc("POST /api/nod", handleJSON(rm, tc))
	rm.Public.HandleFunc("POST /api/form/nod", handleForm(rm, tc))
}
