package templates

import (
	"net/http"

	admintemplates "chameth.com/chameth.com/features/admin/templates"
)

var listPastesTemplate = admintemplates.ParsePage(listPastesGotpl)
var editPasteTemplate = admintemplates.ParsePage(editPasteGotpl)

type ListPastesData = admintemplates.ListData[PasteSummary]

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
