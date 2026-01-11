package templates

import (
	"html/template"
	"net/http"
)

var editFilmListTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"edit-film-list.html.gotpl",
		),
)

type EditFilmListData struct {
	PageData
	ID          int
	Title       string
	Description string
	Published   bool
	Path        string
	Entries     []FilmListEntryItem
	Films       []FilmOption
}

type FilmListEntryItem struct {
	EntryID  int
	FilmID   int
	Position int
	Title    string
	Year     string
}

type FilmOption struct {
	ID    int
	Title string
	Year  string
}

func RenderEditFilmList(w http.ResponseWriter, data EditFilmListData) error {
	return editFilmListTemplate.Execute(w, data)
}
