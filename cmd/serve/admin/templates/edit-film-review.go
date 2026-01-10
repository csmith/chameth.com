package templates

import (
	"html/template"
	"net/http"
)

var editFilmReviewTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"edit-film-review.html.gotpl",
		),
)

type EditFilmReviewData struct {
	PageData
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
