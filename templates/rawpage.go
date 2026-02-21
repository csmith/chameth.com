package templates

import (
	"html/template"
	"net/http"
)

var rawPageTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"rawpage.html.gotpl",
		),
)

type RawPageData struct {
	PageData
	RawContent template.HTML
}

func RenderRawPage(w http.ResponseWriter, page RawPageData) error {
	return rawPageTemplate.Execute(w, page)
}
