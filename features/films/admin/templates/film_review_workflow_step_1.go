package templates

import (
	_ "embed"
	"html/template"
	"net/http"

	admintemplates "chameth.com/chameth.com/features/admin/templates"
)

type FilmBasic struct {
	ID     int
	TMDBID int
	Title  string
	Year   *int
	Path   string
}

type Step1Data struct {
	Films []FilmBasic
}

//go:embed film-review-workflow-step-1.html.gotpl
var filmReviewWorkflowStep1Gotpl string

var filmReviewWorkflowStep1Template = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(filmReviewWorkflowStep1Gotpl))
	return t
}()

func RenderFilmReviewWorkflowStep1(w http.ResponseWriter, data Step1Data) error {
	return filmReviewWorkflowStep1Template.Execute(w, data)
}
