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

	var renderedReviews []template.HTML
	for _, review := range reviews {
		md, err := markdown.Render(review.ReviewText)
		if err != nil {
			return "", fmt.Errorf("failed to render film review markdown: %w", err)
		}

		stars, err := rating.Render(review.Rating)
		if err != nil {
			return "", fmt.Errorf("failed to render film review rating: %w", err)
		}

		replacement, err := filmreview.Render(filmreview.Data{
			Name:       review.Title,
			PosterPath: review.Poster.Path,
			Stars:      template.HTML(stars),
			Rating:     review.Rating,
			Date:       review.WatchedDate.Format("2006-01-02"),
			Rewatch:    review.IsRewatch,
			Spoiler:    review.HasSpoilers,
			Review:     md,
		})
		if err != nil {
			return "", fmt.Errorf("failed to render film review: %w", err)
		}

		renderedReviews = append(renderedReviews, template.HTML(replacement))
	}

	return renderTemplate(Data{
		Reviews: renderedReviews,
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
