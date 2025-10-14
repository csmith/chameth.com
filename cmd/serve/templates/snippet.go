package templates

import (
	"html/template"
	"net/http"
)

var snippetTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"snippet.html.gotpl",
		),
)

type SnippetData struct {
	PageData
	SnippetTitle   string
	SnippetGroup   string
	SnippetContent template.HTML
}

func RenderSnippet(w http.ResponseWriter, snippet SnippetData) error {
	return snippetTemplate.Execute(w, snippet)
}
