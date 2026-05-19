package templates

import (
	_ "embed"
	"html/template"
	"net/http"

	admintemplates "chameth.com/chameth.com/features/admin/templates"
)

type Step6Data struct {
	FilmID int
	Film   FilmBasic
}

//go:embed film-review-workflow-step-6.html.gotpl
var filmReviewWorkflowStep6Gotpl string

var filmReviewWorkflowStep6Template = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(filmReviewWorkflowStep6Gotpl))
	return t
}()

func RenderFilmReviewWorkflowStep6(w http.ResponseWriter, data Step6Data) error {
	return filmReviewWorkflowStep6Template.Execute(w, data)
}
