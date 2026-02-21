package templates

import (
	"html/template"
	"net/http"
)

type Step6Data struct {
	FilmID int
	Film   FilmBasic
}

var filmReviewWorkflowStep6Template = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"film-review-workflow-step-6.html.gotpl",
		),
)

func RenderFilmReviewWorkflowStep6(w http.ResponseWriter, data Step6Data) error {
	return filmReviewWorkflowStep6Template.Execute(w, data)
}
