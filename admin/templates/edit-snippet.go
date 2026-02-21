package templates

import (
	"html/template"
	"net/http"
)

var editSnippetTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"edit-snippet.html.gotpl",
		),
)

type EditSnippetData struct {
	PageData
	ID              int
	Path            string
	Title           string
	Topic           string
	Content         string
	Published       bool
	AvailableTopics []string
}

func RenderEditSnippet(w http.ResponseWriter, data EditSnippetData) error {
	return editSnippetTemplate.Execute(w, data)
}
