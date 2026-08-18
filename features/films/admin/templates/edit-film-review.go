package templates

import (
	_ "embed"
	"net/http"

	admintemplates "chameth.com/chameth.com/features/admin/templates"
)

//go:embed edit-film-review.html.gotpl
var editFilmReviewGotpl string

var editFilmReviewTemplate = admintemplates.ParsePage(editFilmReviewGotpl)

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
