package templates

import (
	"net/http"

	admintemplates "chameth.com/chameth.com/features/admin/templates"
)

var listSyndicationsTemplate = admintemplates.ParsePage(listSyndicationsGotpl)

var editSyndicationTemplate = admintemplates.ParsePage(editSyndicationGotpl)

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
	Disposition string
	Rel         string
}

type EditSyndicationData struct {
	admintemplates.PageData
	ID          int
	Path        string
	ExternalURL string
	Name        string
	Published   bool
	Disposition string
	Rel         string
}

func RenderListSyndications(w http.ResponseWriter, data ListSyndicationsData) error {
	return listSyndicationsTemplate.Execute(w, data)
}

func RenderEditSyndication(w http.ResponseWriter, data EditSyndicationData) error {
	return editSyndicationTemplate.Execute(w, data)
}
