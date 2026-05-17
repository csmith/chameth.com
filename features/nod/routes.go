package nod

import (
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/nod", handleJSON)
	mux.HandleFunc("POST /api/form/nod", handleForm)
}
