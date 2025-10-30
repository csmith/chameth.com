package templates

import (
	"html/template"
	"net/http"
)

var listPoemsTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"list-poems.html.gotpl",
		),
)

type ListPoemsData struct {
	PageData
	Drafts []PoemSummary
	Poems  []PoemSummary
}

type PoemSummary struct {
	ID    int
	Path  string
	Title string
	Date  string
}

func RenderListPoems(w http.ResponseWriter, data ListPoemsData) error {
	return listPoemsTemplate.Execute(w, data)
}
