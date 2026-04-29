package templates

import (
	_ "embed"
	"html/template"
	"net/http"

	admintemplates "chameth.com/chameth.com/admin/templates"
)

type Step3Data struct {
	FilmID            int
	Film              FilmBasic
	LetterboxdListURL string
	Position          int
}

//go:embed film-review-workflow-step-3.html.gotpl
var filmReviewWorkflowStep3Gotpl string

var filmReviewWorkflowStep3Template = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(filmReviewWorkflowStep3Gotpl))
	return t
}()

func RenderFilmReviewWorkflowStep3(w http.ResponseWriter, data Step3Data) error {
	return filmReviewWorkflowStep3Template.Execute(w, data)
}
