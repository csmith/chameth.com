package templates

import (
	_ "embed"
	"html/template"
)

//go:embed page.html.gotpl
var pageGotpl string

// PageData is the base data type for admin pages rendered via ParsePage.
type PageData struct{}

// ListData is the standard data shape for admin list pages: draft items and
// all items, both mapped to summaries.
type ListData[S any] struct {
	PageData
	Drafts []S
	Items  []S
}

// ParsePage parses content over a fresh copy of the standard admin page
// template. The content must define the "content" block.
func ParsePage(content string) *template.Template {
	t := template.Must(template.New("page.html.gotpl").Parse(pageGotpl))
	return template.Must(t.Parse(content))
}

// ParsePageWithFuncs is ParsePage with custom template functions, which must
// be registered before the content is parsed.
func ParsePageWithFuncs(funcs template.FuncMap, content string) *template.Template {
	t := template.Must(template.New("page.html.gotpl").Funcs(funcs).Parse(pageGotpl))
	return template.Must(t.Parse(content))
}
