package shortcodes

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"chameth.com/chameth.com/cmd/serve/content/markdown"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/templates"
	"chameth.com/chameth.com/cmd/serve/db"
)

var (
	filmReviewRegexp = regexp.MustCompile(`\{%\s*filmreview ([0-9]+)\s*%}`)
)

func renderFilmReview(input string) (string, error) {
	res := input
	reviews := filmReviewRegexp.FindAllStringSubmatch(input, -1)
	for _, review := range reviews {
		reviewID := review[1]

		id, err := strconv.Atoi(reviewID)
		if err != nil {
			return "", fmt.Errorf("invalid film review ID: %s", reviewID)
		}

		reviewData, err := db.GetFilmReviewWithFilmAndPoster(id)
		if err != nil {
			return "", fmt.Errorf("failed to get film review: %w", err)
		}

		md, err := markdown.Render(reviewData.ReviewText)
		if err != nil {
			return "", fmt.Errorf("failed to render film review markdown: %w", err)
		}

		replacement, err := templates.RenderFilmReview(templates.FilmReviewData{
			Name:       reviewData.Title,
			PosterPath: reviewData.Poster.Path,
			Rating:     reviewData.Rating,
			Date:       reviewData.WatchedDate,
			Rewatch:    reviewData.IsRewatch,
			Spoiler:    reviewData.HasSpoilers,
			Review:     md,
		})
		if err != nil {
			return "", fmt.Errorf("failed to render film review template: %w", err)
		}

		res = strings.Replace(res, review[0], replacement, 1)
	}
	return res, nil
}
