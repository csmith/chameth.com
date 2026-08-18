package templates

import (
	"net/http"

	admintemplates "chameth.com/chameth.com/features/admin/templates"
)

var listPoemsTemplate = admintemplates.ParsePage(listPoemsGotpl)
var editPoemTemplate = admintemplates.ParsePage(editPoemGotpl)

type ListPoemsData = admintemplates.ListData[PoemSummary]

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
