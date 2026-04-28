package templates

import (
	"html/template"
	"net/http"
)

var editGoImportTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			Templates,
			"page.html.gotpl",
			"edit-goimport.html.gotpl",
		),
)

type EditGoImportData struct {
	PageData
	ID        int
	Path      string
	VCS       string
	RepoURL   string
	Published bool
}

func RenderEditGoImport(w http.ResponseWriter, data EditGoImportData) error {
	return editGoImportTemplate.Execute(w, data)
}
