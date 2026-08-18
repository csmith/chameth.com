package templates

import (
	_ "embed"
	"net/http"

	admintemplates "chameth.com/chameth.com/features/admin/templates"
)

type Step3Data struct {
	FilmID            int
	Film              FilmBasic
	LetterboxdListURL string
	Position          int
}

//go:embed film-review-workflow-step-3.html.gotpl
var filmReviewWorkflowStep3Gotpl string

var filmReviewWorkflowStep3Template = admintemplates.ParsePage(filmReviewWorkflowStep3Gotpl)

func RenderFilmReviewWorkflowStep3(w http.ResponseWriter, data Step3Data) error {
	return filmReviewWorkflowStep3Template.Execute(w, data)
}
