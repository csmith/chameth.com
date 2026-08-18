package templates

import (
	_ "embed"
	"net/http"

	admintemplates "chameth.com/chameth.com/features/admin/templates"
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

//go:embed film-review-workflow-step-7.html.gotpl
var filmReviewWorkflowStep7Gotpl string

var filmReviewWorkflowStep7Template = admintemplates.ParsePage(filmReviewWorkflowStep7Gotpl)

func RenderFilmReviewWorkflowStep7(w http.ResponseWriter, data Step7Data) error {
	return filmReviewWorkflowStep7Template.Execute(w, data)
}
