package templates

import (
	"html/template"
	"net/http"
)

var goImportTemplate = template.Must(
	template.
		New("go.html.gotpl").
		ParseFS(
			templates,
			"go.html.gotpl",
		),
)

type GoImportData struct {
	ModulePath string
	VCS        string
	RepoURL    string
}

func RenderGoImport(w http.ResponseWriter, data GoImportData) error {
	return goImportTemplate.Execute(w, data)
}
