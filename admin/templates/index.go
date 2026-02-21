package templates

import (
	"html/template"
	"net/http"
)

var indexTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"index.html.gotpl",
		),
)

type IndexData struct {
	PageData
}

func RenderIndex(w http.ResponseWriter, data IndexData) error {
	return indexTemplate.Execute(w, data)
}
