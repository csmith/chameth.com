package templates

import (
	"html/template"
	"net/http"
)

var siteMapTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"sitemap.html.gotpl",
		),
)

type SiteMapData struct {
	PageData
	Poems    []ContentDetails
	Posts    []ContentDetails
	Snippets []SnippetDetails
}

type ContentDetails struct {
	Title string
	Url   string
	Date  ContentDate
}

type ContentDate struct {
	Iso      string
	Friendly string
}

func RenderSiteMap(w http.ResponseWriter, data SiteMapData) error {
	return siteMapTemplate.ExecuteTemplate(w, "page.html.gotpl", data)
}
