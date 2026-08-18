package templates

import (
	_ "embed"
	"net/http"

	admintemplates "chameth.com/chameth.com/features/admin/templates"
)

//go:embed list-film-lists.html.gotpl
var listFilmListsGotpl string

var listFilmListsTemplate = admintemplates.ParsePage(listFilmListsGotpl)

type ListFilmListsData struct {
	admintemplates.PageData
	Drafts []FilmListSummary
	Lists  []FilmListSummary
}

type FilmListSummary struct {
	ID        int
	Title     string
	Path      string
	Published bool
}

func RenderListFilmLists(w http.ResponseWriter, data ListFilmListsData) error {
	return listFilmListsTemplate.Execute(w, data)
}
