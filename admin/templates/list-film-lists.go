package templates

import (
	"html/template"
	"net/http"
)

var listFilmListsTemplate = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"list-film-lists.html.gotpl",
		),
)

type ListFilmListsData struct {
	PageData
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
