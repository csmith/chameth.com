package templates

import (
	"html/template"
	"io"
)

var projectsTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"projects.html.gotpl",
		),
)

type ProjectsData struct {
	PageData
	ProjectGroups []ProjectGroup
}

type ProjectGroup struct {
	Name        string
	Description string
	Projects    []ProjectDetails
}

type ProjectDetails struct {
	Name        string
	Pinned      bool
	Icon        template.HTML
	Description template.HTML
}

func RenderProjects(w io.Writer, projectsData ProjectsData) error {
	return projectsTemplate.Execute(w, projectsData)
}
