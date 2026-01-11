package templates

import (
	"html/template"
	"net/http"
)

var editFilmTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"edit-film.html.gotpl",
		),
)

type EditFilmData struct {
	PageData
	ID        int
	Title     string
	Year      string
	Overview  string
	Runtime   string
	Published bool
	Path      string
	Poster    *MediaItem
	Reviews   []FilmReviewSummary
}

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
