package posters

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"chameth.com/chameth.com/features/films"
	"chameth.com/chameth.com/features/shortcodes/rating"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("posters.html.gotpl").ParseFS(templates, "posters.html.gotpl"))

type Layout string

const (
	Strip Layout = "recent-films"
	Grid  Layout = "watched-films"
)

type Film struct {
	Title      string
	PosterPath string
	Path       string
	Date       string
	Stars      template.HTML
}

type Data struct {
	Layout Layout
	Films  []Film
}

func Render(layout Layout, entries []films.ReviewEntry) (string, error) {
	filmList := make([]Film, len(entries))
	for i, entry := range entries {
		stars, err := rating.Render(entry.Rating)
		if err != nil {
			return "", fmt.Errorf("failed to render film rating: %w", err)
		}

		filmList[i] = Film{
			Title:      entry.FilmTitle,
			PosterPath: entry.PosterPath,
			Path:       entry.FilmPath,
			Date:       entry.WatchedDate,
			Stars:      template.HTML(stars),
		}
	}

	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, Data{Layout: layout, Films: filmList}); err != nil {
		return "", err
	}
	return buf.String(), nil
}
