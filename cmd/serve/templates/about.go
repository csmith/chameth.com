package templates

import (
	"html/template"
	"net/http"

	"chameth.com/chameth.com/cmd/serve/templates/includes"
)

var aboutTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"about.html.gotpl",
			"includes/postlink.html.gotpl",
		),
)

type AboutData struct {
	PageData
	HighlightedPosts []includes.PostLinkData
}

func RenderAbout(w http.ResponseWriter, aboutData AboutData) error {
	return aboutTemplate.Execute(w, aboutData)
}
