package shortcodes

import (
	"fmt"
	"html/template"
	"regexp"
	"strings"

	"chameth.com/chameth.com/cmd/serve/content/markdown"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/templates"
	"chameth.com/chameth.com/cmd/serve/db"
)

var (
	filmReviewsRegexp = regexp.MustCompile(`\{%\s*filmreviews\s*%}`)
)

func renderFilmReviews(input string) (string, error) {
	res := input
	matches := filmReviewsRegexp.FindAllStringSubmatch(input, -1)

	if len(matches) == 0 {
		return res, nil
	}

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

		replacement, err := templates.RenderFilmReview(templates.FilmReviewData{
			Name:       review.Title,
			PosterPath: review.Poster.Path,
			Rating:     review.Rating,
			Date:       review.WatchedDate,
			Rewatch:    review.IsRewatch,
			Spoiler:    review.HasSpoilers,
			Review:     md,
		})
		if err != nil {
			return "", fmt.Errorf("failed to render film review template: %w", err)
		}

		renderedReviews = append(renderedReviews, template.HTML(replacement))
	}

	filmReviewsHTML, err := templates.RenderFilmReviews(templates.FilmReviewsData{
		Reviews: renderedReviews,
	})
	if err != nil {
		return "", fmt.Errorf("failed to render film reviews template: %w", err)
	}

	for _, match := range matches {
		res = strings.Replace(res, match[0], filmReviewsHTML, 1)
	}

	return res, nil
}
