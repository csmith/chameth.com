package templates

import (
	"html/template"
	"net/http"
)

var atomTemplate = template.Must(
	template.
		New("atom.xml.gotpl").
		ParseFS(
			templates,
			"atom.xml.gotpl",
		),
)

type AtomData struct {
	FeedTitle       string
	FeedLastUpdated string
	FeedItems       []FeedItem
}

type FeedItem struct {
	Title   string
	Link    string
	Updated string
	Content string
}

func RenderAtom(w http.ResponseWriter, data AtomData) error {
	return atomTemplate.Execute(w, data)
}
