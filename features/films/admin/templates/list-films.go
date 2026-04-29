package templates

import (
	_ "embed"
	"html/template"
	"net/http"

	admintemplates "chameth.com/chameth.com/admin/templates"
)

//go:embed list-films.html.gotpl
var listFilmsGotpl string

var listFilmsTemplate = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(listFilmsGotpl))
	return t
}()

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
