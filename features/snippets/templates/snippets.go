package templates

import (
	"fmt"
	"html/template"
	"io"

	parenttemplates "chameth.com/chameth.com/templates"
)

var snippetTemplate = func() *template.Template {
	pageContent, err := parenttemplates.FS.ReadFile("page.html.gotpl")
	if err != nil {
		panic(fmt.Sprintf("failed to read page.html.gotpl: %v", err))
	}
	t := template.Must(template.New("page.html.gotpl").Parse(string(pageContent)))
	template.Must(t.Parse(snippetTemplateContent))
	return t
}()

var snippetsListTemplate = func() *template.Template {
	pageContent, err := parenttemplates.FS.ReadFile("page.html.gotpl")
	if err != nil {
		panic(fmt.Sprintf("failed to read page.html.gotpl: %v", err))
	}
	t := template.Must(template.New("page.html.gotpl").Parse(string(pageContent)))
	template.Must(t.Parse(snippetsTemplateContent))
	return t
}()

type SnippetData struct {
	parenttemplates.PageData
	SnippetTitle   string
	SnippetGroup   string
	SnippetContent template.HTML
}

type SnippetsData struct {
	parenttemplates.PageData
	SnippetGroups []SnippetGroup
}

type SnippetGroup struct {
	Name     string
	Snippets []SnippetDetails
}

type SnippetDetails struct {
	Name string
	Path string
}

func RenderSnippet(w io.Writer, data SnippetData) error {
	return snippetTemplate.Execute(w, data)
}

func RenderSnippets(w io.Writer, data SnippetsData) error {
	return snippetsListTemplate.Execute(w, data)
}
