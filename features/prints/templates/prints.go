package templates

import (
	"fmt"
	"html/template"
	"io"

	parenttemplates "chameth.com/chameth.com/templates"
)

var printsTemplate = func() *template.Template {
	pageContent, err := parenttemplates.FS.ReadFile("page.html.gotpl")
	if err != nil {
		panic(fmt.Sprintf("failed to read page.html.gotpl: %v", err))
	}
	t := template.Must(template.New("page.html.gotpl").Parse(string(pageContent)))

	template.Must(t.Parse(printsTemplateContent))
	return t
}()

type PrintsData struct {
	parenttemplates.PageData
	Prints []PrintDetails
}

type PrintDetails struct {
	Name        string
	Description string
	RenderPath  string
	PreviewPath string
	Links       []PrintLink
}

type PrintLink struct {
	Name    string
	Address string
}

func RenderPrints(w io.Writer, data PrintsData) error {
	return printsTemplate.Execute(w, data)
}
