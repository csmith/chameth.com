package templates

import (
	"html/template"
	"net/http"
	textTemplate "text/template"
)

var siteMapHtmlTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"sitemap.html.gotpl",
		),
)

var siteMapXmlTemplate = textTemplate.Must(
	textTemplate.
		New("sitemap.xml.gotpl").
		ParseFS(
			templates,
			"sitemap.xml.gotpl",
		),
)

type SiteMapData struct {
	PageData
	Poems     []ContentDetails
	Posts     []ContentDetails
	Snippets  []SnippetDetails
	Films     []ContentDetails
	FilmLists []ContentDetails
}

type ContentDetails struct {
	Title string
	Path  string
	Date  ContentDate
}

type ContentDate struct {
	Iso      string
	Friendly string
}

func RenderHtmlSiteMap(w http.ResponseWriter, data SiteMapData) error {
	return siteMapHtmlTemplate.ExecuteTemplate(w, "page.html.gotpl", data)
}

func RenderXmlSiteMap(w http.ResponseWriter, data SiteMapData) error {
	return siteMapXmlTemplate.ExecuteTemplate(w, "sitemap.xml.gotpl", data)
}
