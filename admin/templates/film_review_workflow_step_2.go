package templates

import (
	"html/template"
	"net/http"

	"chameth.com/chameth.com/db"
)

type Step2Data struct {
	FilmID      int
	Film        FilmBasic
	Entries     []FilmListEntryWithPoster
	EndPosition int
}

type FilmListEntryWithPoster struct {
	Entry         db.FilmListEntry
	Film          FilmBasic
	PosterMediaID *int
	AverageRating int
	RatingHTML    template.HTML
}

var filmReviewWorkflowStep2Template = template.Must(
	template.
		New("page.html.gotpl").
		Funcs(template.FuncMap{
			"add": func(a, b int) int { return a + b },
		}).
		ParseFS(
			templates,
			"page.html.gotpl",
			"film-review-workflow-step-2.html.gotpl",
		),
)

func RenderFilmReviewWorkflowStep2(w http.ResponseWriter, data Step2Data) error {
	return filmReviewWorkflowStep2Template.Execute(w, data)
}
