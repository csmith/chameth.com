package templates

import (
	"html/template"
	"net/http"
)

type Step4Data struct {
	FilmID        int
	Film          FilmBasic
	WatchedDate   string
	DefaultRating int
}

var filmReviewWorkflowStep4Template = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"film-review-workflow-step-4.html.gotpl",
		),
)

func RenderFilmReviewWorkflowStep4(w http.ResponseWriter, data Step4Data) error {
	return filmReviewWorkflowStep4Template.Execute(w, data)
}
