package templates

import (
	"html/template"
	"net/http"

	"github.com/csmith/chameth.com/cmd/serve/templates/includes"
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
	Interests        AboutInterests
}

type AboutInterests struct {
	Languages  []string
	VideoGames []string
	BoardGames []string
	Books      []string
	Films      []string
}

func RenderAbout(w http.ResponseWriter, aboutData AboutData) error {
	return aboutTemplate.Execute(w, aboutData)
}
