package templates

import (
	"html/template"
	"net/http"
)

var editMediaRelationsTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"edit-media-relations.html.gotpl",
		),
)

type EditMediaRelationsData struct {
	PageData
	EntityType     string
	EntityID       int
	EntitySlug     string
	Media          []MediaRelationItem
	AvailableMedia []AvailableMediaItem
}

type MediaRelationItem struct {
	Slug        string
	Title       string
	AltText     string
	Width       *int
	Height      *int
	Role        string
	ContentType string
	MediaID     int
	IsVariant   bool
}

type AvailableMediaItem struct {
	MediaID          int
	OriginalFilename string
	ContentType      string
	Width            *int
	Height           *int
	IsVariant        bool
}

func RenderEditMediaRelations(w http.ResponseWriter, data EditMediaRelationsData) error {
	return editMediaRelationsTemplate.Execute(w, data)
}
