package shortcodes

import (
	"fmt"
	"strconv"

	"chameth.com/chameth.com/cmd/serve/content/shortcodes/templates"
)

func renderRating(args []string, _ *Context) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("rating requires at least 1 argument (value)")
	}

	numericRating, err := strconv.Atoi(args[0])
	if err != nil {
		return "", fmt.Errorf("invalid rating %s: %w", args[0], err)
	}

	return RenderRating(numericRating)
}

func RenderRating(rating int) (string, error) {
	return templates.RenderRating(templates.RatingData{
		FilledStars: rating / 2,
		HalfStar:    rating%2 == 1,
		EmptyStars:  5 - ((rating + 1) / 2),
	})
}
