package templates

import (
	"html/template"
	"net/http"
)

var notFoundTemplate = template.Must(
	template.
		New("page.html.gotpl").
		Funcs(funcMap).
		ParseFS(
			templates,
			"page.html.gotpl",
			"notfound.html.gotpl",
		),
)

type NotFoundData struct {
	PageData
}

func RenderNotFound(w http.ResponseWriter, data NotFoundData) error {
	return notFoundTemplate.Execute(w, data)
}
