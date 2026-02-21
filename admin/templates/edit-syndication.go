package templates

import (
	"html/template"
	"net/http"
)

var editSyndicationTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"edit-syndication.html.gotpl",
		),
)

type EditSyndicationData struct {
	PageData
	ID          int
	Path        string
	ExternalURL string
	Name        string
	Published   bool
}

func RenderEditSyndication(w http.ResponseWriter, data EditSyndicationData) error {
	return editSyndicationTemplate.Execute(w, data)
}
