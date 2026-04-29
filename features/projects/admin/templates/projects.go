package templates

import (
	"html/template"
	"net/http"

	admintemplates "chameth.com/chameth.com/admin/templates"
)

var listProjectsTemplate = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(listProjectsGotpl))
	return t
}()

var editProjectTemplate = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(editProjectGotpl))
	return t
}()

type ListProjectsData struct {
	admintemplates.PageData
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
