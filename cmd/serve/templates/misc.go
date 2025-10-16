package templates

import (
	"html/template"
	"net/http"
)

var miscTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"misc.html.gotpl",
		),
)

type MiscData struct {
	PageData
	Poems []PoemDetails
}

type PoemDetails struct {
	Title string
	Url   string
}

func RenderMisc(w http.ResponseWriter, data MiscData) error {
	return miscTemplate.ExecuteTemplate(w, "page.html.gotpl", data)
}
