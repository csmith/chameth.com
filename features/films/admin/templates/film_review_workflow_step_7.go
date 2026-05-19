package templates

import (
	_ "embed"
	"html/template"
	"net/http"

	admintemplates "chameth.com/chameth.com/features/admin/templates"
)

type FilmListWithLetterboxd struct {
	ID                int
	Title             string
	Path              string
	LetterboxdListURL string
}

type Step7Data struct {
	FilmID   int
	Film     FilmBasic
	AllLists []FilmListWithLetterboxd
}

//go:embed film-review-workflow-step-7.html.gotpl
var filmReviewWorkflowStep7Gotpl string

var filmReviewWorkflowStep7Template = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(filmReviewWorkflowStep7Gotpl))
	return t
}()

func RenderFilmReviewWorkflowStep7(w http.ResponseWriter, data Step7Data) error {
	return filmReviewWorkflowStep7Template.Execute(w, data)
}
