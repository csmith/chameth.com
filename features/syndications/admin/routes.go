package admin

import (
	"chameth.com/chameth.com/features/routing"
)

func RegisterRoutes(rm *routing.Manager) {
	rm.Admin.HandleFunc("GET /syndications", ListSyndicationsHandler())
	rm.Admin.HandleFunc("POST /syndications/create", CreateSyndicationHandler())
	rm.Admin.HandleFunc("GET /syndications/edit/{id}", EditSyndicationHandler())
	rm.Admin.HandleFunc("POST /syndications/edit/{id}", UpdateSyndicationHandler())
	rm.Admin.HandleFunc("POST /syndications/delete/{id}", DeleteSyndicationHandler())
}
