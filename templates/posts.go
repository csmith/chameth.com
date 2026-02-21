package templates

import (
	"html/template"
	"net/http"
)

var postsTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"posts.html.gotpl",
		),
)

type PostsData struct {
	PageData
	Posts []string
}

func RenderPosts(w http.ResponseWriter, postsData PostsData) error {
	return postsTemplate.Execute(w, postsData)
}
