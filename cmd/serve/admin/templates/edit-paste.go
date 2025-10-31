package templates

import (
	"html/template"
	"net/http"
)

var editPasteTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"edit-paste.html.gotpl",
		),
)

type EditPasteData struct {
	PageData
	ID        int
	Path      string
	Title     string
	Language  string
	Content   string
	Date      string
	Published bool
}

func RenderEditPaste(w http.ResponseWriter, data EditPasteData) error {
	return editPasteTemplate.Execute(w, data)
}
