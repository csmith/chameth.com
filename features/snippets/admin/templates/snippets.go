package templates

import (
	"net/http"

	admintemplates "chameth.com/chameth.com/features/admin/templates"
)

var listSnippetsTemplate = admintemplates.ParsePage(listSnippetsGotpl)
var editSnippetTemplate = admintemplates.ParsePage(editSnippetGotpl)

type ListSnippetsData = admintemplates.ListData[SnippetSummary]

type SnippetSummary struct {
	ID    int
	Path  string
	Title string
	Topic string
}

type EditSnippetData struct {
	admintemplates.PageData
	ID              int
	Path            string
	Title           string
	Topic           string
	Content         string
	Published       bool
	AvailableTopics []string
}

func RenderListSnippets(w http.ResponseWriter, data ListSnippetsData) error {
	return listSnippetsTemplate.Execute(w, data)
}

func RenderEditSnippet(w http.ResponseWriter, data EditSnippetData) error {
	return editSnippetTemplate.Execute(w, data)
}
