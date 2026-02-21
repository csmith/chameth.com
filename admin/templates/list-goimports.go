package templates

import (
	"html/template"
	"net/http"
)

var listGoImportsTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"list-goimports.html.gotpl",
		),
)

type ListGoImportsData struct {
	PageData
	Drafts    []GoImportSummary
	GoImports []GoImportSummary
}

type GoImportSummary struct {
	ID      int
	Path    string
	VCS     string
	RepoURL string
}

func RenderListGoImports(w http.ResponseWriter, data ListGoImportsData) error {
	return listGoImportsTemplate.Execute(w, data)
}
