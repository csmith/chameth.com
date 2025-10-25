package templates

import (
	"html/template"
	"net/http"
)

var editPoemTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"edit-poem.html.gotpl",
		),
)

type EditPoemData struct {
	PageData
	ID        int
	Slug      string
	Title     string
	Poem      string
	Notes     string
	Date      string
	Published bool
}

func RenderEditPoem(w http.ResponseWriter, data EditPoemData) error {
	return editPoemTemplate.Execute(w, data)
}
