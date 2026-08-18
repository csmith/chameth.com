package templates

import (
	_ "embed"
	"net/http"

	admintemplates "chameth.com/chameth.com/features/admin/templates"
)

type Step6Data struct {
	FilmID int
	Film   FilmBasic
}

//go:embed film-review-workflow-step-6.html.gotpl
var filmReviewWorkflowStep6Gotpl string

var filmReviewWorkflowStep6Template = admintemplates.ParsePage(filmReviewWorkflowStep6Gotpl)

func RenderFilmReviewWorkflowStep6(w http.ResponseWriter, data Step6Data) error {
	return filmReviewWorkflowStep6Template.Execute(w, data)
}
