package contact

import (
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/contact", handleJSON)
	mux.HandleFunc("POST /api/form/contact", handleForm)
}
