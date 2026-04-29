package templates

import (
	"html/template"
	"net/http"

	admintemplates "chameth.com/chameth.com/admin/templates"
)

var listPastesTemplate = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(listPastesGotpl))
	return t
}()

var editPasteTemplate = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(editPasteGotpl))
	return t
}()

type ListPastesData struct {
	admintemplates.PageData
	Drafts []PasteSummary
	Pastes []PasteSummary
}

type PasteSummary struct {
	ID       int
	Path     string
	Title    string
	Language string
}

type EditPasteData struct {
	admintemplates.PageData
	ID        int
	Path      string
	Title     string
	Language  string
	Content   string
	Date      string
	Published bool
}

func RenderListPastes(w http.ResponseWriter, data ListPastesData) error {
	return listPastesTemplate.Execute(w, data)
}

func RenderEditPaste(w http.ResponseWriter, data EditPasteData) error {
	return editPasteTemplate.Execute(w, data)
}
