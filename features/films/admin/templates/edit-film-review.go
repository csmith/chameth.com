package templates

import (
	_ "embed"
	"html/template"
	"net/http"

	admintemplates "chameth.com/chameth.com/admin/templates"
)

//go:embed edit-film-review.html.gotpl
var editFilmReviewGotpl string

var editFilmReviewTemplate = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(editFilmReviewGotpl))
	return t
}()

type EditFilmReviewData struct {
	admintemplates.PageData
	ID          int
	FilmID      int
	FilmTitle   string
	WatchedDate string
	Rating      string
	IsRewatch   bool
	HasSpoilers bool
	ReviewText  string
	Published   bool
}

func RenderEditFilmReview(w http.ResponseWriter, data EditFilmReviewData) error {
	return editFilmReviewTemplate.Execute(w, data)
}
