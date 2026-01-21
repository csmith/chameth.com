package templates

import (
	"html/template"
	"net/http"
)

var listSyndicationsTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"list-syndications.html.gotpl",
		),
)

type ListSyndicationsData struct {
	PageData
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

func RenderListSyndications(w http.ResponseWriter, data ListSyndicationsData) error {
	return listSyndicationsTemplate.Execute(w, data)
}
