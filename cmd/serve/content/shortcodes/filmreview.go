package shortcodes

import (
	"fmt"
	"html/template"
	"strconv"

	"chameth.com/chameth.com/cmd/serve/content/markdown"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/templates"
	"chameth.com/chameth.com/cmd/serve/db"
)

func renderFilmReview(args []string, _ *Context) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("filmreview requires at least 1 argument (id)")
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return "", fmt.Errorf("invalid film review ID: %s", args[0])
	}

	return RenderFilmReview(id)
}

func RenderFilmReview(id int) (string, error) {
	reviewData, err := db.GetFilmReviewWithFilmAndPoster(id)
	if err != nil {
		return "", fmt.Errorf("failed to get film review: %w", err)
	}

	md, err := markdown.Render(reviewData.ReviewText)
	if err != nil {
		return "", fmt.Errorf("failed to render film review markdown: %w", err)
	}

	stars, err := RenderRating(reviewData.Rating)
	if err != nil {
		return "", fmt.Errorf("failed to render film review stars: %w", err)
	}

	replacement, err := templates.RenderFilmReview(templates.FilmReviewData{
		Name:       reviewData.Title,
		PosterPath: reviewData.Poster.Path,
		Rating:     reviewData.Rating,
		Stars:      template.HTML(stars),
		Date:       reviewData.WatchedDate,
		Rewatch:    reviewData.IsRewatch,
		Spoiler:    reviewData.HasSpoilers,
		Review:     md,
	})
	if err != nil {
		return "", fmt.Errorf("failed to render film review template: %w", err)
	}
	return replacement, nil
}
