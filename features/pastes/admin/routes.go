package admin

import (
	"chameth.com/chameth.com/features/routing"
)

func RegisterRoutes(rm *routing.Manager) {
	rm.Admin.HandleFunc("GET /pastes", ListPastesHandler())
	rm.Admin.HandleFunc("POST /pastes", CreatePasteHandler())
	rm.Admin.HandleFunc("GET /pastes/edit/{id}", EditPasteHandler())
	rm.Admin.HandleFunc("POST /pastes/edit/{id}", UpdatePasteHandler())
}
