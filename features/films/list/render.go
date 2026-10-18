package list

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"strconv"

	"chameth.com/chameth.com/content/markdown"
	"chameth.com/chameth.com/features/films"
	"chameth.com/chameth.com/features/shortcodes"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("filmlist.html.gotpl").ParseFS(templates, "filmlist.html.gotpl"))

func RenderFromText(args []string, ctx *shortcodes.Context) (string, error) {
	if len(args) < 1 {
		return "", errors.New("filmlist requires at least 1 argument (id)")
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return "", fmt.Errorf("invalid film list ID: %s", args[0])
	}

	return Render(ctx.Context, id)
}

func Render(ctx context.Context, id int) (string, error) {
	card, err := films.GetFilmListCard(ctx, id, 5)
	if err != nil {
		return "", fmt.Errorf("failed to get film list: %w", err)
	}

	filmList := make([]Film, len(card.Films))
	for i, film := range card.Films {
		year := ""
		if film.Year != nil {
			year = strconv.Itoa(*film.Year)
		}

		posterPath := ""
		if film.PosterPath != nil {
			posterPath = *film.PosterPath
		}

		filmList[i] = Film{
			ID:         film.ID,
			Title:      film.Title,
			Year:       year,
			PosterPath: posterPath,
			Path:       film.Path,
		}
	}

	description, err := markdown.Render(card.Description)
	if err != nil {
		return "", fmt.Errorf("failed to render film list description: %w", err)
	}

	return renderTemplate(
		Data{
			ID:          card.ID,
			Title:       card.Title,
			Description: description,
			Path:        card.Path,
			Count:       card.Count,
			Films:       filmList,
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
