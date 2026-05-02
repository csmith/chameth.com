package templates

import (
	"html/template"
	"net/http"

	admintemplates "chameth.com/chameth.com/admin/templates"
)

var listSyndicationsTemplate = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(listSyndicationsGotpl))
	return t
}()

var editSyndicationTemplate = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(editSyndicationGotpl))
	return t
}()

type ListSyndicationsData struct {
	admintemplates.PageData
	Unpublished  []SyndicationSummary
	Syndications []SyndicationSummary
}

type SyndicationSummary struct {
	ID          int
	Path        string
	ExternalURL string
	Name        string
	Published   bool
}

type EditSyndicationData struct {
	admintemplates.PageData
	ID          int
	Path        string
	ExternalURL string
	Name        string
	Published   bool
}

func RenderListSyndications(w http.ResponseWriter, data ListSyndicationsData) error {
	return listSyndicationsTemplate.Execute(w, data)
}

func RenderEditSyndication(w http.ResponseWriter, data EditSyndicationData) error {
	return editSyndicationTemplate.Execute(w, data)
}
