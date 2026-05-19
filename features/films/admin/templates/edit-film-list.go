package templates

import (
	_ "embed"
	"html/template"
	"net/http"

	admintemplates "chameth.com/chameth.com/features/admin/templates"
)

//go:embed edit-film-list.html.gotpl
var editFilmListGotpl string

var editFilmListTemplate = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(editFilmListGotpl))
	return t
}()

type EditFilmListData struct {
	admintemplates.PageData
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
