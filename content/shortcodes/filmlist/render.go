package filmlist

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"strconv"

	"chameth.com/chameth.com/content/markdown"
	"chameth.com/chameth.com/content/shortcodes/common"
	"chameth.com/chameth.com/db"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("filmlist.html.gotpl").ParseFS(templates, "filmlist.html.gotpl"))

func RenderFromText(args []string, ctx *common.Context) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("filmlist requires at least 1 argument (id)")
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return "", fmt.Errorf("invalid film list ID: %s", args[0])
	}

	return Render(ctx.Context, id)
}

func Render(ctx context.Context, id int) (string, error) {
	list, count, entries, err := db.GetFilmListWithFilms(ctx, id)
	if err != nil {
		return "", fmt.Errorf("failed to get film list: %w", err)
	}

	films := make([]Film, len(entries))
	for i, entry := range entries {
		year := ""
		if entry.Film.Year != nil {
			year = strconv.Itoa(*entry.Film.Year)
		}

		posterPath := ""
		if entry.Poster.Path != "" {
			posterPath = entry.Poster.Path
		}

		films[i] = Film{
			ID:         entry.Film.ID,
			Title:      entry.Film.Title,
			Year:       year,
			PosterPath: posterPath,
			Path:       entry.Film.Path,
		}
	}

	description, err := markdown.Render(list.Description)
	if err != nil {
		return "", fmt.Errorf("failed to render film list description: %w", err)
	}

	return renderTemplate(Data{
		ID:          list.ID,
		Title:       list.Title,
		Description: description,
		Path:        list.Path,
		Count:       count,
		Films:       films,
	})
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
