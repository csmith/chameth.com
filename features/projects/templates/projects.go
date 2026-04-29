package templates

import (
	"fmt"
	"html/template"
	"io"

	parenttemplates "chameth.com/chameth.com/templates"
)

var projectsTemplate = func() *template.Template {
	pageContent, err := parenttemplates.FS.ReadFile("page.html.gotpl")
	if err != nil {
		panic(fmt.Sprintf("failed to read page.html.gotpl: %v", err))
	}
	t := template.Must(template.New("page.html.gotpl").Parse(string(pageContent)))

	template.Must(t.Parse(projectsTemplateContent))
	return t
}()

type ProjectsData struct {
	parenttemplates.PageData
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

func RenderProjects(w io.Writer, data ProjectsData) error {
	return projectsTemplate.Execute(w, data)
}
