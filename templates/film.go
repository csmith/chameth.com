package templates

import (
	"html/template"
	"io"
)

var filmTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"film.html.gotpl",
		),
)

type FilmData struct {
	PageData
	Title         string
	Year          string
	TMDBID        *int
	Overview      template.HTML
	Reviews       []FilmReviewData
	TimesWatched  int
	AverageRating int
	PosterPath    string
	FilmLists     []int
}

type FilmReviewData struct {
	WatchedDate string
	Rating      int
	IsRewatch   bool
	HasSpoilers bool
	Content     template.HTML
}

func RenderFilm(w io.Writer, film FilmData) error {
	return filmTemplate.Execute(w, film)
}
