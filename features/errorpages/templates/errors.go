package templates

import (
	"fmt"
	"html/template"
	"io"

	parenttemplates "chameth.com/chameth.com/templates"
)

var notFoundTemplate = func() *template.Template {
	pageContent, err := parenttemplates.FS.ReadFile("page.html.gotpl")
	if err != nil {
		panic(fmt.Sprintf("failed to read page.html.gotpl: %v", err))
	}
	t := template.Must(template.New("page.html.gotpl").Parse(string(pageContent)))
	template.Must(t.Parse(notFoundTemplateContent))
	return t
}()

var serverErrorTemplate = func() *template.Template {
	pageContent, err := parenttemplates.FS.ReadFile("page.html.gotpl")
	if err != nil {
		panic(fmt.Sprintf("failed to read page.html.gotpl: %v", err))
	}
	t := template.Must(template.New("page.html.gotpl").Parse(string(pageContent)))
	template.Must(t.Parse(serverErrorTemplateContent))
	return t
}()

type NotFoundData struct {
	parenttemplates.PageData
}

type ServerErrorData struct {
	parenttemplates.PageData
}

func RenderNotFound(w io.Writer, data NotFoundData) error {
	return notFoundTemplate.Execute(w, data)
}

func RenderServerError(w io.Writer, data ServerErrorData) error {
	return serverErrorTemplate.Execute(w, data)
}
