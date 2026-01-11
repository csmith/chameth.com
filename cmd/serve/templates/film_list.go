package templates

import (
	"html/template"
	"io"
)

var filmListTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"film_list.html.gotpl",
			"includes/postlink.html.gotpl",
		),
)

type FilmListData struct {
	PageData
	Title       string
	Description template.HTML
	Entries     []FilmListItem
}

type FilmListItem struct {
	Position     int
	PosterPath   string
	FilmPath     string
	Title        string
	Year         string
	TimesWatched int
	RatingText   string
	RatingHTML   template.HTML
	LastWatched  string
}

func RenderFilmList(w io.Writer, filmList FilmListData) error {
	return filmListTemplate.Execute(w, filmList)
}
