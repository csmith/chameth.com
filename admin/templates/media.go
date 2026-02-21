package templates

import (
	"html/template"
	"net/http"
)

var mediaTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"media.html.gotpl",
		),
)

type MediaData struct {
	PageData
	MediaItems []MediaItem
}

type MediaItem struct {
	ID               int
	OriginalFilename string
	ParentMediaID    *int
	Width            *int
	Height           *int
	ContentType      string
}

func RenderMedia(w http.ResponseWriter, data MediaData) error {
	return mediaTemplate.Execute(w, data)
}
