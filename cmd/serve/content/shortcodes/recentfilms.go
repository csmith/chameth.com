package shortcodes

import (
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"strings"

	"chameth.com/chameth.com/cmd/serve/content/shortcodes/templates"
	"chameth.com/chameth.com/cmd/serve/db"
)

var (
	recentFilmsRegexp = regexp.MustCompile(`\{%\s*recentfilms ([0-9]+)\s*%}`)
)

func renderRecentFilms(input string, _ *Context) (string, error) {
	res := input
	matches := recentFilmsRegexp.FindAllStringSubmatch(input, -1)
	for _, match := range matches {
		countStr := match[1]

		count, err := strconv.Atoi(countStr)
		if err != nil {
			return "", fmt.Errorf("invalid recent films count: %s", countStr)
		}

		replacement, err := RenderRecentFilms(count)
		if err != nil {
			return "", err
		}

		res = strings.Replace(res, match[0], replacement, 1)
	}
	return res, nil
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
