package filmreviews

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"chameth.com/chameth.com/cmd/serve/content/markdown"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/context"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/filmreview"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/rating"
	"chameth.com/chameth.com/cmd/serve/db"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("filmreviews.html.gotpl").ParseFS(templates, "filmreviews.html.gotpl"))

func RenderFromText(args []string, _ *context.Context) (string, error) {
	reviews, err := db.GetAllPublishedFilmReviewsWithFilmAndPosters()
	if err != nil {
		return "", fmt.Errorf("failed to get film reviews: %w", err)
	}

	var html []template.HTML
	for _, data := range reviews {
		md, err := markdown.Render(data.ReviewText)
		if err != nil {
			return "", fmt.Errorf("failed to render film review markdown: %w", err)
		}

		stars, err := rating.Render(data.Rating)
		if err != nil {
			return "", fmt.Errorf("failed to render film review rating: %w", err)
		}

		replacement, err := filmreview.Render(filmreview.Data{
			Name:       data.Film.Title,
			Path:       data.Film.Path,
			PosterPath: data.Poster.Path,
			Stars:      template.HTML(stars),
			Rating:     data.FilmReview.Rating,
			Date:       data.FilmReview.WatchedDate.Format("2006-01-02"),
			Rewatch:    data.FilmReview.IsRewatch,
			Spoiler:    data.FilmReview.HasSpoilers,
			Review:     md,
		})
		if err != nil {
			return "", fmt.Errorf("failed to render film review: %w", err)
		}

		html = append(html, template.HTML(replacement))
	}

	return renderTemplate(Data{
		Reviews: html,
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
