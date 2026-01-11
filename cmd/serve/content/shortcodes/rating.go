package shortcodes

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"chameth.com/chameth.com/cmd/serve/content/shortcodes/templates"
)

var (
	ratingRegexp = regexp.MustCompile(`(?s)\{%\s*rating ([0-9]+)\s*%}`)
)

func renderRating(input string) (string, error) {
	res := input
	ratings := ratingRegexp.FindAllStringSubmatch(input, -1)
	for _, rating := range ratings {
		numericRating, err := strconv.Atoi(rating[1])
		if err != nil {
			return "", fmt.Errorf("invalid rating %s: %w", rating[1], err)
		}

		replacement, err := RenderRating(numericRating)
		if err != nil {
			return "", fmt.Errorf("failed to render sidenote template: %w", err)
		}

		res = strings.Replace(res, rating[0], replacement, 1)
	}
	return res, nil
}

func RenderRating(rating int) (string, error) {
	return templates.RenderRating(templates.RatingData{
		FilledStars: rating / 2,
		HalfStar:    rating%2 == 1,
		EmptyStars:  5 - ((rating + 1) / 2),
	})
}
