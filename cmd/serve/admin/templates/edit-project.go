package templates

import (
	"html/template"
	"net/http"
)

var editProjectTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"edit-project.html.gotpl",
		),
)

type EditProjectData struct {
	PageData
	ID                int
	Name              string
	Icon              string
	Description       string
	Section           int
	Pinned            bool
	Published         bool
	AvailableSections []SectionOption
}

type SectionOption struct {
	ID   int
	Name string
}

func RenderEditProject(w http.ResponseWriter, data EditProjectData) error {
	return editProjectTemplate.Execute(w, data)
}
