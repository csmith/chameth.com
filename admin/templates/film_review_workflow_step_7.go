package templates

import (
	"html/template"
	"net/http"
)

type FilmListWithLetterboxd struct {
	ID                int
	Title             string
	Path              string
	LetterboxdListURL string
}

type Step7Data struct {
	FilmID   int
	Film     FilmBasic
	AllLists []FilmListWithLetterboxd
}

var filmReviewWorkflowStep7Template = template.Must(
	template.
		New("page.html.gotpl").
		ParseFS(
			templates,
			"page.html.gotpl",
			"film-review-workflow-step-7.html.gotpl",
		),
)

func RenderFilmReviewWorkflowStep7(w http.ResponseWriter, data Step7Data) error {
	return filmReviewWorkflowStep7Template.Execute(w, data)
}
