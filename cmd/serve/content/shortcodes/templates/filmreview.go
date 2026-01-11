package templates

import (
	"bytes"
	"html/template"
	"time"
)

var filmReviewTemplate = template.Must(
	template.
		New("filmreview.html.gotpl").
		Funcs(template.FuncMap{
			"formatDate": func(t time.Time) string {
				return t.Format("2006-01-02")
			},
		}).
		ParseFS(
			templates,
			"filmreview.html.gotpl",
		),
)

type FilmReviewData struct {
	Name       string
	PosterPath string
	Rating     int
	Stars      template.HTML
	Date       time.Time
	Rewatch    bool
	Spoiler    bool
	Review     template.HTML
}

func RenderFilmReview(data FilmReviewData) (string, error) {
	buf := &bytes.Buffer{}
	err := filmReviewTemplate.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
