package templates

import (
	_ "embed"
	"net/http"

	admintemplates "chameth.com/chameth.com/features/admin/templates"
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

var filmReviewWorkflowStep5Template = admintemplates.ParsePage(filmReviewWorkflowStep5Gotpl)

func RenderFilmReviewWorkflowStep5(w http.ResponseWriter, data Step5Data) error {
	return filmReviewWorkflowStep5Template.Execute(w, data)
}
