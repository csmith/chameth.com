package templates

import (
	_ "embed"
	"net/http"

	admintemplates "chameth.com/chameth.com/features/admin/templates"
	mediatemplates "chameth.com/chameth.com/features/media/admin/templates"
)

//go:embed edit-film.html.gotpl
var editFilmGotpl string

var editFilmTemplate = admintemplates.ParsePage(editFilmGotpl)

type EditFilmData struct {
	admintemplates.PageData
	ID        int
	Title     string
	Year      string
	TMDBID    *int
	Overview  string
	Runtime   string
	Published bool
	Path      string
	Poster    *mediatemplates.MediaItem
	Reviews   []FilmReviewSummary
}

type MediaItem = mediatemplates.MediaItem

type FilmReviewSummary struct {
	ID          int
	WatchedDate string
	Rating      string
	IsRewatch   bool
	HasSpoilers bool
	ReviewText  string
	Published   bool
}

type SearchResult struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	Year       string `json:"year"`
	PosterPath string `json:"poster_path"`
	Overview   string `json:"overview"`
}

func RenderEditFilm(w http.ResponseWriter, data EditFilmData) error {
	return editFilmTemplate.Execute(w, data)
}
