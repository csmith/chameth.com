package filmreview

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strconv"

	"chameth.com/chameth.com/cmd/serve/content/markdown"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/context"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/rating"
	"chameth.com/chameth.com/cmd/serve/db"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("filmreview.html.gotpl").ParseFS(templates, "filmreview.html.gotpl"))

func RenderFromText(args []string, _ *context.Context) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("filmreview requires at least 1 argument (id)")
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return "", fmt.Errorf("invalid film review ID: %s", args[0])
	}

	data, err := db.GetFilmReviewWithFilmAndPoster(id)
	if err != nil {
		return "", fmt.Errorf("failed to get film review: %w", err)
	}

	md, err := markdown.Render(data.FilmReview.ReviewText)
	if err != nil {
		return "", fmt.Errorf("failed to render film review markdown: %w", err)
	}

	stars, err := rating.Render(data.FilmReview.Rating)
	if err != nil {
		return "", fmt.Errorf("failed to render film review stars: %w", err)
	}

	return renderTemplate(Data{
		Name:       data.Film.Title,
		Path:       data.Film.Path,
		PosterPath: data.Poster.Path,
		Rating:     data.FilmReview.Rating,
		Stars:      template.HTML(stars),
		Date:       data.FilmReview.WatchedDate.Format("2006-01-02"),
		Rewatch:    data.FilmReview.IsRewatch,
		Spoiler:    data.FilmReview.HasSpoilers,
		Review:     md,
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

func Render(data Data) (string, error) {
	return renderTemplate(data)
}
