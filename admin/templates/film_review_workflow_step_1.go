package templates

import (
	"html/template"
	"net/http"
)

type FilmBasic struct {
	ID     int
	TMDBID int
	Title  string
	Year   *int
	Path   string
}

type Step1Data struct {
	Films []FilmBasic
}

var filmReviewWorkflowStep1Template = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"film-review-workflow-step-1.html.gotpl",
		),
)

func RenderFilmReviewWorkflowStep1(w http.ResponseWriter, data Step1Data) error {
	return filmReviewWorkflowStep1Template.Execute(w, data)
}
