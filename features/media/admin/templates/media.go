package templates

import (
	"html/template"
	"net/http"

	admintemplates "chameth.com/chameth.com/admin/templates"
)

var mediaTemplate = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(mediaGotpl))
	return t
}()

var editMediaTemplate = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(editMediaGotpl))
	return t
}()

var editMediaRelationsTemplate = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(editMediaRelationsGotpl))
	return t
}()

type MediaItem struct {
	ID               int
	OriginalFilename string
	ParentMediaID    *int
	Width            *int
	Height           *int
	ContentType      string
}

type MediaData struct {
	admintemplates.PageData
	MediaItems []MediaItem
}

type EditMediaData struct {
	admintemplates.PageData
	ID               int
	OriginalFilename string
	ParentMediaID    *int
	Width            *int
	Height           *int
	ContentType      string
}

type EditMediaRelationsData struct {
	admintemplates.PageData
	EntityType     string
	EntityID       int
	EntityPath     string
	Media          []MediaRelationItem
	AvailableMedia []AvailableMediaItem
}

type MediaRelationItem struct {
	Path        string
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

func RenderMedia(w http.ResponseWriter, data MediaData) error {
	return mediaTemplate.Execute(w, data)
}

func RenderEditMedia(w http.ResponseWriter, data EditMediaData) error {
	return editMediaTemplate.Execute(w, data)
}

func RenderEditMediaRelations(w http.ResponseWriter, data EditMediaRelationsData) error {
	return editMediaRelationsTemplate.Execute(w, data)
}
