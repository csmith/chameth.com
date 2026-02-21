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
	Path      string
	Date      string
	Content   string
	Format    string
	Published bool
	Media     []PostMediaItem
}

type PostMediaItem struct {
	Path        string
	Title       string
	AltText     string
	Width       *int
	Height      *int
	Role        string
	ContentType string
	MediaID     int
	Variants    []PostMediaVariant
}

type PostMediaVariant struct {
	MediaID     int
	ContentType string
	Width       *int
	Height      *int
}

func RenderEditPost(w http.ResponseWriter, data EditPostData) error {
	return editPostTemplate.Execute(w, data)
}
