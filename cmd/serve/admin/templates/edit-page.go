package templates

import (
	"html/template"
	"net/http"
)

var editPageTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"edit-page.html.gotpl",
		),
)

type EditPageData struct {
	PageData
	ID        int
	Title     string
	Path      string
	Content   string
	Published bool
	Raw       bool
	Media     []PageMediaItem
}

type PageMediaItem struct {
	Path        string
	Title       string
	AltText     string
	Width       *int
	Height      *int
	Role        string
	ContentType string
	MediaID     int
	Variants    []PageMediaVariant
}

type PageMediaVariant struct {
	MediaID     int
	ContentType string
	Width       *int
	Height      *int
}

func RenderEditPage(w http.ResponseWriter, data EditPageData) error {
	return editPageTemplate.Execute(w, data)
}
