package templates

import (
	"html/template"
	"net/http"

	"github.com/csmith/chameth.com/cmd/serve/templates/includes"
)

var postsTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"posts.html.gotpl",
			"includes/postlink.html.gotpl",
		),
)

type PostsData struct {
	PageData
	Posts []includes.PostLinkData
}

func RenderPosts(w http.ResponseWriter, postsData PostsData) error {
	return postsTemplate.Execute(w, postsData)
}
