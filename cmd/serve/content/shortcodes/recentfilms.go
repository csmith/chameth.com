package shortcodes

import (
	"fmt"
	"html/template"
	"strconv"

	"chameth.com/chameth.com/cmd/serve/content/shortcodes/templates"
	"chameth.com/chameth.com/cmd/serve/db"
)

func renderRecentFilms(args []string, _ *Context) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("recentfilms requires at least 1 argument (count)")
	}

	count, err := strconv.Atoi(args[0])
	if err != nil {
		return "", fmt.Errorf("invalid recent films count: %s", args[0])
	}

	return RenderRecentFilms(count)
}

func RenderRecentFilms(n int) (string, error) {
	reviews, err := db.GetRecentPublishedFilmReviewsWithFilmAndPosters(n)
	if err != nil {
		return "", fmt.Errorf("failed to get recent film reviews: %w", err)
	}

	films := make([]templates.RecentFilmData, len(reviews))
	for i, review := range reviews {
		stars, err := RenderRating(review.Rating)
		if err != nil {
			return "", fmt.Errorf("failed to render film rating: %w", err)
		}

		posterPath := ""
		if review.Poster.Path != "" {
			posterPath = review.Poster.Path
		}

		films[i] = templates.RecentFilmData{
			Title:      review.Title,
			PosterPath: posterPath,
			Path:       review.Path,
			Date:       review.WatchedDate,
			Stars:      template.HTML(stars),
		}
	}

	return templates.RenderRecentFilms(templates.RecentFilmsData{
		Films: films,
	})
}
