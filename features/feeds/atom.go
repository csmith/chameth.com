package feeds

import (
	"embed"
	"html"
	"text/template"
	"io"
)

//go:embed atom.xml.gotpl
var atomTemplateFS embed.FS

var atomTemplate = template.Must(
	template.
		New("atom.xml.gotpl").
		Funcs(template.FuncMap{
			"escape": html.EscapeString,
		}).
		ParseFS(
			atomTemplateFS,
			"atom.xml.gotpl",
		),
)

type AtomData struct {
	FeedTitle       string
	FeedSelfLink    string
	FeedLastUpdated string
	FeedItems       []FeedItem
}

type FeedItem struct {
	Title   string
	Link    string
	Updated string
	Content string
}

func renderAtom(w io.Writer, data AtomData) error {
	return atomTemplate.Execute(w, data)
}
