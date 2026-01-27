package templates

import (
	"chameth.com/chameth.com/cmd/serve/db"
	"html/template"
	"net/http"
)

type Step5Data struct {
	FilmID            int
	Film              FilmBasic
	ReviewID          int
	Review            db.FilmReview
	LetterboxdFilmURL string
}

var filmReviewWorkflowStep5Template = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"film-review-workflow-step-5.html.gotpl",
		),
)

func RenderFilmReviewWorkflowStep5(w http.ResponseWriter, data Step5Data) error {
	return filmReviewWorkflowStep5Template.Execute(w, data)
}
