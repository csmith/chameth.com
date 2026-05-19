package templates

import (
	_ "embed"
	"html/template"
	"net/http"

	admintemplates "chameth.com/chameth.com/features/admin/templates"
	"chameth.com/chameth.com/features/films"
)

type Step2Data struct {
	FilmID      int
	Film        FilmBasic
	Entries     []FilmListEntryWithPoster
	EndPosition int
}

type FilmListEntryWithPoster struct {
	Entry         films.FilmListEntry
	Film          FilmBasic
	PosterMediaID *int
	AverageRating int
	RatingHTML    template.HTML
}

//go:embed film-review-workflow-step-2.html.gotpl
var filmReviewWorkflowStep2Gotpl string

var filmReviewWorkflowStep2Template = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}).Parse(string(page)))
	template.Must(t.Parse(filmReviewWorkflowStep2Gotpl))
	return t
}()

func RenderFilmReviewWorkflowStep2(w http.ResponseWriter, data Step2Data) error {
	return filmReviewWorkflowStep2Template.Execute(w, data)
}
