package sitemap

import (
	"embed"
	"html/template"
	"io"
	textTemplate "text/template"

	parenttemplates "chameth.com/chameth.com/templates"
)

//go:embed sitemap.html.gotpl sitemap.xml.gotpl
var templateFS embed.FS

var siteMapHtmlTemplate = func() *template.Template {
	t := template.Must(
		template.
			New("page.html.gotpl").
			ParseFS(
				parenttemplates.FS,
				"page.html.gotpl",
			),
	)
	return template.Must(t.ParseFS(
		templateFS,
		"sitemap.html.gotpl",
	))
}()

var siteMapXmlTemplate = textTemplate.Must(
	textTemplate.
		New("sitemap.xml.gotpl").
		ParseFS(
			templateFS,
			"sitemap.xml.gotpl",
		),
)

type SiteMapData struct {
	parenttemplates.PageData
	Poems     []parenttemplates.ContentDetails
	Posts     []parenttemplates.ContentDetails
	Snippets  []parenttemplates.ContentDetails
	Films     []parenttemplates.ContentDetails
	FilmLists []parenttemplates.ContentDetails
	Pages     []SiteMapPageDetails
}

type SiteMapPageDetails struct {
	Title       string
	Path        string
	Frequency   string
	Priority    string
	CurrentPage bool
}

func renderHtmlSiteMap(w io.Writer, data SiteMapData) error {
	return siteMapHtmlTemplate.ExecuteTemplate(w, "page.html.gotpl", data)
}

func renderXmlSiteMap(w io.Writer, data SiteMapData) error {
	return siteMapXmlTemplate.ExecuteTemplate(w, "sitemap.xml.gotpl", data)
}
