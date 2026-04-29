package templates

import (
	"fmt"
	"html/template"
	"io"

	parenttemplates "chameth.com/chameth.com/templates"
)

var staticPageTemplate = func() *template.Template {
	pageContent, err := parenttemplates.FS.ReadFile("page.html.gotpl")
	if err != nil {
		panic(fmt.Sprintf("failed to read page.html.gotpl: %v", err))
	}
	t := template.Must(template.New("page.html.gotpl").Parse(string(pageContent)))
	template.Must(t.Parse(staticPageTemplateContent))
	return t
}()

var rawPageTemplate = func() *template.Template {
	pageContent, err := parenttemplates.FS.ReadFile("page.html.gotpl")
	if err != nil {
		panic(fmt.Sprintf("failed to read page.html.gotpl: %v", err))
	}
	t := template.Must(template.New("page.html.gotpl").Parse(string(pageContent)))
	template.Must(t.Parse(rawPageTemplateContent))
	return t
}()

type StaticPageData struct {
	parenttemplates.PageData
	StaticTitle   string
	StaticContent template.HTML
}

type RawPageData struct {
	parenttemplates.PageData
	RawContent template.HTML
}

func RenderStaticPage(w io.Writer, data StaticPageData) error {
	return staticPageTemplate.Execute(w, data)
}

func RenderRawPage(w io.Writer, data RawPageData) error {
	return rawPageTemplate.Execute(w, data)
}
