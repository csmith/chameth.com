package templates

import (
	"html/template"
	"net/http"

	admintemplates "chameth.com/chameth.com/features/admin/templates"
)

var listProjectsTemplate = admintemplates.ParsePage(listProjectsGotpl)
var editProjectTemplate = admintemplates.ParsePage(editProjectGotpl)

type ListProjectsData = admintemplates.ListData[ProjectSummary]

type ProjectSummary struct {
	ID          int
	Name        string
	Icon        template.HTML
	Pinned      bool
	Section     string
	Description string
	SectionSort int
}

type EditProjectData struct {
	admintemplates.PageData
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

func RenderListProjects(w http.ResponseWriter, data ListProjectsData) error {
	return listProjectsTemplate.Execute(w, data)
}

func RenderEditProject(w http.ResponseWriter, data EditProjectData) error {
	return editProjectTemplate.Execute(w, data)
}
