package templates

import (
	"html/template"
	"net/http"
)

var listPostsTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"list-posts.html.gotpl",
		),
)

type ListPostsData struct {
	PageData
	Drafts []PostSummary
	Posts  []PostSummary
}

type PostSummary struct {
	ID    int
	Title string
	Path  string
	Date  string
}

func RenderListPosts(w http.ResponseWriter, data ListPostsData) error {
	return listPostsTemplate.Execute(w, data)
}
