package templates

import (
	"bytes"
	"html/template"
	"time"
)

var recentFilmsTemplate = template.Must(
	template.
		New("recentfilms.html.gotpl").
		Funcs(template.FuncMap{
			"formatDate": func(t time.Time) string {
				return t.Format("2006-01-02")
			},
		}).
		ParseFS(
			templates,
			"recentfilms.html.gotpl",
		),
)

type RecentFilmData struct {
	Title      string
	PosterPath string
	Path       string
	Date       time.Time
	Stars      template.HTML
}

type RecentFilmsData struct {
	Films []RecentFilmData
}

func RenderRecentFilms(data RecentFilmsData) (string, error) {
	buf := &bytes.Buffer{}
	err := recentFilmsTemplate.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
