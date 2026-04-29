package templates

import (
	_ "embed"
	"html/template"
	"net/http"

	admintemplates "chameth.com/chameth.com/admin/templates"
	"chameth.com/chameth.com/features/films"
)

type Step5Data struct {
	FilmID            int
	Film              FilmBasic
	ReviewID          int
	Review            films.FilmReview
	LetterboxdFilmURL string
}

//go:embed film-review-workflow-step-5.html.gotpl
var filmReviewWorkflowStep5Gotpl string

var filmReviewWorkflowStep5Template = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(filmReviewWorkflowStep5Gotpl))
	return t
}()

func RenderFilmReviewWorkflowStep5(w http.ResponseWriter, data Step5Data) error {
	return filmReviewWorkflowStep5Template.Execute(w, data)
}
