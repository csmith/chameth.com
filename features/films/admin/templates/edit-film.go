package templates

import (
	_ "embed"
	"html/template"
	"net/http"

	admintemplates "chameth.com/chameth.com/admin/templates"
)

//go:embed edit-film.html.gotpl
var editFilmGotpl string

var editFilmTemplate = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(editFilmGotpl))
	return t
}()

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
	Poster    *admintemplates.MediaItem
	Reviews   []FilmReviewSummary
}

type MediaItem = admintemplates.MediaItem

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
