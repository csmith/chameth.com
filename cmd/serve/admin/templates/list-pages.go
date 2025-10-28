package templates

import (
	"html/template"
	"net/http"
)

var listPagesTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"list-pages.html.gotpl",
		),
)

type ListPagesData struct {
	PageData
	Drafts []PageSummary
	Pages  []PageSummary
}

type PageSummary struct {
	ID    int
	Title string
	Slug  string
}

func RenderListPages(w http.ResponseWriter, data ListPagesData) error {
	return listPagesTemplate.Execute(w, data)
}
