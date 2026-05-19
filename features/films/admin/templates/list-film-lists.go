package templates

import (
	_ "embed"
	"html/template"
	"net/http"

	admintemplates "chameth.com/chameth.com/features/admin/templates"
)

//go:embed list-film-lists.html.gotpl
var listFilmListsGotpl string

var listFilmListsTemplate = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(listFilmListsGotpl))
	return t
}()

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
