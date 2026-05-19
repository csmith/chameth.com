package admin

import (
	"chameth.com/chameth.com/features/routing"
)

func RegisterRoutes(rm *routing.Manager) {
	rm.Admin.HandleFunc("GET /media", MediaHandler())
	rm.Admin.HandleFunc("POST /media/upload", UploadMediaHandler())
	rm.Admin.HandleFunc("GET /media/view/{id}", ViewMediaHandler())
	rm.Admin.HandleFunc("GET /media/edit/{id}", EditMediaHandler())
	rm.Admin.HandleFunc("POST /media/edit/{id}", ReplaceMediaHandler())
	rm.Admin.HandleFunc("GET /media-relations/edit", EditMediaRelationsHandler())
	rm.Admin.HandleFunc("POST /media-relations/update", UpdateMediaRelationHandler())
	rm.Admin.HandleFunc("POST /media-relations/remove", RemoveMediaRelationHandler())
	rm.Admin.HandleFunc("POST /media-relations/add", AddMediaRelationsHandler())
}
