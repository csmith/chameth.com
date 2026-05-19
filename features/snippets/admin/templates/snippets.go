package templates

import (
	"html/template"
	"net/http"

	admintemplates "chameth.com/chameth.com/features/admin/templates"
)

var listSnippetsTemplate = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(listSnippetsGotpl))
	return t
}()

var editSnippetTemplate = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(editSnippetGotpl))
	return t
}()

type ListSnippetsData struct {
	admintemplates.PageData
	Drafts   []SnippetSummary
	Snippets []SnippetSummary
}

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
