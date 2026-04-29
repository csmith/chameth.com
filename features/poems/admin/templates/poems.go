package templates

import (
	"html/template"
	"net/http"

	admintemplates "chameth.com/chameth.com/admin/templates"
)

var listPoemsTemplate = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(listPoemsGotpl))
	return t
}()

var editPoemTemplate = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(editPoemGotpl))
	return t
}()

type ListPoemsData struct {
	admintemplates.PageData
	Drafts []PoemSummary
	Poems  []PoemSummary
}

type PoemSummary struct {
	ID    int
	Path  string
	Title string
	Date  string
}

type EditPoemData struct {
	admintemplates.PageData
	ID        int
	Path      string
	Title     string
	Poem      string
	Notes     string
	Date      string
	Published bool
}

func RenderListPoems(w http.ResponseWriter, data ListPoemsData) error {
	return listPoemsTemplate.Execute(w, data)
}

func RenderEditPoem(w http.ResponseWriter, data EditPoemData) error {
	return editPoemTemplate.Execute(w, data)
}
