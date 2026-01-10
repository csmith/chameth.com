package templates

import (
	"html/template"
	"net/http"
)

var listFilmsTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"list-films.html.gotpl",
		),
)

type ListFilmsData struct {
	PageData
	Films []FilmSummary
}

type FilmSummary struct {
	ID        int
	Title     string
	Year      string
	Rating    string
	Published bool
}

func RenderListFilms(w http.ResponseWriter, data ListFilmsData) error {
	return listFilmsTemplate.Execute(w, data)
}
