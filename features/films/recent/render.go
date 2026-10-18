package recent

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"chameth.com/chameth.com/features/films"
	"chameth.com/chameth.com/features/shortcodes"
	"chameth.com/chameth.com/features/shortcodes/rating"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("recentfilms.html.gotpl").ParseFS(templates, "recentfilms.html.gotpl"))

func render(_ []string, entries []films.ReviewEntry, _ *shortcodes.Context) (string, error) {
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

	return renderTemplate(
		Data{
			Films: filmList,
		},
	)
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
