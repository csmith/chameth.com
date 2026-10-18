package review

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"

	"chameth.com/chameth.com/content/markdown"
	"chameth.com/chameth.com/features/films"
	"chameth.com/chameth.com/features/shortcodes"
	"chameth.com/chameth.com/features/shortcodes/rating"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("filmreview.html.gotpl").ParseFS(templates, "filmreview.html.gotpl"))

func render(_ []string, entry *films.ReviewEntry, _ *shortcodes.Context) (string, error) {
	if entry == nil {
		return "", errors.New("film review not found")
	}
	return RenderEntry(*entry)
}

func RenderEntry(entry films.ReviewEntry) (string, error) {
	md, err := markdown.Render(entry.Text)
	if err != nil {
		return "", fmt.Errorf("failed to render film review markdown: %w", err)
	}

	stars, err := rating.Render(entry.Rating)
	if err != nil {
		return "", fmt.Errorf("failed to render film review stars: %w", err)
	}

	return Render(
		Data{
			Name:       entry.FilmTitle,
			Path:       entry.FilmPath,
			PosterPath: entry.PosterPath,
			Rating:     entry.Rating,
			Stars:      template.HTML(stars),
			Date:       entry.WatchedDate,
			Rewatch:    entry.IsRewatch,
			Spoiler:    entry.HasSpoilers,
			Review:     md,
		},
	)
}

func Render(data Data) (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
