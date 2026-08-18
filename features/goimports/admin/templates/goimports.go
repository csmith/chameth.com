package templates

import (
	"net/http"

	admintemplates "chameth.com/chameth.com/features/admin/templates"
)

var listGoImportsTemplate = admintemplates.ParsePage(listGoImportsGotpl)
var editGoImportTemplate = admintemplates.ParsePage(editGoImportGotpl)

type ListGoImportsData = admintemplates.ListData[GoImportSummary]

type GoImportSummary struct {
	ID      int
	Path    string
	VCS     string
	RepoURL string
}

type EditGoImportData struct {
	admintemplates.PageData
	ID        int
	Path      string
	VCS       string
	RepoURL   string
	Published bool
}

func RenderListGoImports(w http.ResponseWriter, data ListGoImportsData) error {
	return listGoImportsTemplate.Execute(w, data)
}

func RenderEditGoImport(w http.ResponseWriter, data EditGoImportData) error {
	return editGoImportTemplate.Execute(w, data)
}
