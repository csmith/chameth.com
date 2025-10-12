package templates

import (
	"html/template"
	"net/http"
)

var serverErrorTemplate = template.Must(
	template.
		New("page.html.gotpl").
		Funcs(funcMap).
		ParseFS(
			templates,
			"page.html.gotpl",
			"servererror.html.gotpl",
		),
)

type ServerErrorData struct {
	PageData
}

func RenderServerError(w http.ResponseWriter, data ServerErrorData) error {
	return serverErrorTemplate.Execute(w, data)
}
