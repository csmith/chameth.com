package templates

import (
	_ "embed"
	"net/http"

	admintemplates "chameth.com/chameth.com/features/admin/templates"
)

//go:embed list-films.html.gotpl
var listFilmsGotpl string

var listFilmsTemplate = admintemplates.ParsePage(listFilmsGotpl)

type ListFilmsData struct {
	admintemplates.PageData
	Films []FilmSummary
}

type FilmSummary struct {
	ID            int
	Title         string
	Year          string
	Rating        string
	Published     bool
	PosterMediaID *int
	ReviewCount   int
	LastWatched   *string
}

func RenderListFilms(w http.ResponseWriter, data ListFilmsData) error {
	return listFilmsTemplate.Execute(w, data)
}
