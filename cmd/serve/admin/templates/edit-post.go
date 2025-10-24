package templates

import (
	"html/template"
	"net/http"
)

var editPostTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"edit-post.html.gotpl",
		),
)

type EditPostData struct {
	PageData
	ID        int
	Title     string
	Slug      string
	Published string
	Content   string
}

func RenderEditPost(w http.ResponseWriter, data EditPostData) error {
	return editPostTemplate.Execute(w, data)
}
