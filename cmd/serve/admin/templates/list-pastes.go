package templates

import (
	"html/template"
	"net/http"
)

var listPastesTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"list-pastes.html.gotpl",
		),
)

type ListPastesData struct {
	PageData
	Drafts []PasteSummary
	Pastes []PasteSummary
}

type PasteSummary struct {
	ID       int
	Path     string
	Title    string
	Language string
}

func RenderListPastes(w http.ResponseWriter, data ListPastesData) error {
	return listPastesTemplate.Execute(w, data)
}
