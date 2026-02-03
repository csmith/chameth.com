package templates

import (
	"html/template"
	"net/http"
)

type Step3Data struct {
	FilmID            int
	Film              FilmBasic
	LetterboxdListURL string
	Position          int
}

var filmReviewWorkflowStep3Template = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"film-review-workflow-step-3.html.gotpl",
		),
)

func RenderFilmReviewWorkflowStep3(w http.ResponseWriter, data Step3Data) error {
	return filmReviewWorkflowStep3Template.Execute(w, data)
}
