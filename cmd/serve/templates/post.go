package templates

import (
	"html/template"
	"net/http"
)

var postTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"article.html.gotpl",
			"post.html.gotpl",
		),
)

type PostData struct {
	ArticleData
	PostContent template.HTML
	PostFormat  string
	PostTags    []string
}

func RenderPost(w http.ResponseWriter, post PostData) error {
	return postTemplate.Execute(w, post)
}
