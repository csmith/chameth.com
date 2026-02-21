package templates

import (
	"html/template"
	"net/http"
)

var listProjectsTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"list-projects.html.gotpl",
		),
)

type ListProjectsData struct {
	PageData
	Drafts   []ProjectSummary
	Projects []ProjectSummary
}

type ProjectSummary struct {
	ID          int
	Name        string
	Icon        template.HTML
	Pinned      bool
	Section     string
	Description string
	SectionSort int
}

func RenderListProjects(w http.ResponseWriter, data ListProjectsData) error {
	return listProjectsTemplate.Execute(w, data)
}
