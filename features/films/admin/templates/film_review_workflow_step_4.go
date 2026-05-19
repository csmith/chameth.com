package templates

import (
	_ "embed"
	"html/template"
	"net/http"

	admintemplates "chameth.com/chameth.com/features/admin/templates"
)

type Step4Data struct {
	FilmID        int
	Film          FilmBasic
	WatchedDate   string
	DefaultRating int
}

//go:embed film-review-workflow-step-4.html.gotpl
var filmReviewWorkflowStep4Gotpl string

var filmReviewWorkflowStep4Template = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(filmReviewWorkflowStep4Gotpl))
	return t
}()

func RenderFilmReviewWorkflowStep4(w http.ResponseWriter, data Step4Data) error {
	return filmReviewWorkflowStep4Template.Execute(w, data)
}
