package admin

import (
	"chameth.com/chameth.com/features/routing"
)

func RegisterRoutes(rm *routing.Manager) {
	rm.Admin.HandleFunc("GET /projects", ListProjectsHandler())
	rm.Admin.HandleFunc("POST /projects", CreateProjectHandler())
	rm.Admin.HandleFunc("GET /projects/edit/{id}", EditProjectHandler())
	rm.Admin.HandleFunc("POST /projects/edit/{id}", UpdateProjectHandler())
}
