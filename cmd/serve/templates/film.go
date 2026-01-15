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
			"includes/postlink.html.gotpl",
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
	AverageRating template.HTML
	PosterPath    string
}

type FilmReviewData struct {
	WatchedDate string
	Rating      int
	RatingHTML  template.HTML
	IsRewatch   bool
	HasSpoilers bool
	Content     template.HTML
}

func RenderFilm(w io.Writer, film FilmData) error {
	return filmTemplate.Execute(w, film)
}
