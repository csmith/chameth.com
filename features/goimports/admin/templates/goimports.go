package templates

import (
	"html/template"
	"net/http"

	admintemplates "chameth.com/chameth.com/features/admin/templates"
)

var listGoImportsTemplate = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(listGoImportsGotpl))
	return t
}()

var editGoImportTemplate = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(editGoImportGotpl))
	return t
}()

type ListGoImportsData struct {
	admintemplates.PageData
	Drafts    []GoImportSummary
	GoImports []GoImportSummary
}

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
