package templates

import (
	"net/http"

	admintemplates "chameth.com/chameth.com/features/admin/templates"
	"chameth.com/chameth.com/features/media"
)

var listPagesTemplate = admintemplates.ParsePage(listPagesGotpl)
var editPageTemplate = admintemplates.ParsePage(editPageGotpl)

type ListPagesData = admintemplates.ListData[PageSummary]

type PageSummary struct {
	ID    int
	Title string
	Path  string
}

type EditPageData struct {
	admintemplates.PageData
	ID               int
	Title            string
	Path             string
	Content          string
	Published        bool
	Raw              bool
	ParentID         int
	AvailableParents []PageSummary
	SitemapFrequency string
	SitemapPriority  string
	Media            []media.GroupedMedia
}

func RenderListPages(w http.ResponseWriter, data ListPagesData) error {
	return listPagesTemplate.Execute(w, data)
}

func RenderEditPage(w http.ResponseWriter, data EditPageData) error {
	return editPageTemplate.Execute(w, data)
}
