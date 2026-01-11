package templates

import (
	"bytes"
	"html/template"
)

var filmListTemplate = template.Must(
	template.New("filmlist.html.gotpl").ParseFS(
		templates,
		"filmlist.html.gotpl",
	),
)

type FilmListData struct {
	ID          int
	Title       string
	Description template.HTML
	Path        string
	Count       int
	Films       []FilmListFilm
}

type FilmListFilm struct {
	ID         int
	Title      string
	Year       string
	PosterPath string
	Path       string
}

func RenderFilmList(data FilmListData) (string, error) {
	buf := &bytes.Buffer{}
	err := filmListTemplate.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
