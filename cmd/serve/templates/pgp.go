package templates

import (
	"html/template"
	"net/http"
)

var pgpTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"pgp.html.gotpl",
		),
)

type PGPData struct {
	PageData
}

func RenderPGP(w http.ResponseWriter, data PGPData) error {
	return pgpTemplate.Execute(w, data)
}
