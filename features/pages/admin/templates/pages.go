package templates

import (
	"html/template"
	"net/http"

	admintemplates "chameth.com/chameth.com/admin/templates"
)

var listPagesTemplate = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(listPagesGotpl))
	return t
}()

var editPageTemplate = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(editPageGotpl))
	return t
}()

type ListPagesData struct {
	admintemplates.PageData
	Drafts []PageSummary
	Pages  []PageSummary
}

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
	SitemapFrequency string
	SitemapPriority  string
	Media            []PageMediaItem
}

type PageMediaItem struct {
	Path        string
	Title       string
	AltText     string
	Width       *int
	Height      *int
	Role        string
	ContentType string
	MediaID     int
	Variants    []PageMediaVariant
}

type PageMediaVariant struct {
	MediaID     int
	ContentType string
	Width       *int
	Height      *int
}

func RenderListPages(w http.ResponseWriter, data ListPagesData) error {
	return listPagesTemplate.Execute(w, data)
}

func RenderEditPage(w http.ResponseWriter, data EditPageData) error {
	return editPageTemplate.Execute(w, data)
}
