package templates

import (
	"html/template"
	"net/http"
)

var staticPageTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"staticpage.html.gotpl",
		),
)

type StaticPageData struct {
	PageData
	StaticTitle   string
	StaticContent template.HTML
}

func RenderStaticPage(w http.ResponseWriter, page StaticPageData) error {
	return staticPageTemplate.Execute(w, page)
}
