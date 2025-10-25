package templates

import (
	"html/template"
	"net/http"
)

var listSnippetsTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"list-snippets.html.gotpl",
		),
)

type ListSnippetsData struct {
	PageData
	Drafts   []SnippetSummary
	Snippets []SnippetSummary
}

type SnippetSummary struct {
	ID    int
	Slug  string
	Title string
	Topic string
}

func RenderListSnippets(w http.ResponseWriter, data ListSnippetsData) error {
	return listSnippetsTemplate.Execute(w, data)
}
