package templates

import (
	"html/template"
	"net/http"
)

var snippetsTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"snippets.html.gotpl",
		),
)

type SnippetsData struct {
	PageData
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

func RenderSnippets(w http.ResponseWriter, data SnippetsData) error {
	return snippetsTemplate.Execute(w, data)
}
