package templates

import (
	"bytes"
	"html/template"
)

var ratingTemplate = template.Must(
	template.
		New("rating.html.gotpl").
		ParseFS(
			templates,
			"rating.html.gotpl",
		),
)

type RatingData struct {
	FilledStars int
	HalfStar    bool
	EmptyStars  int
}

func RenderRating(data RatingData) (string, error) {
	buf := &bytes.Buffer{}
	err := ratingTemplate.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
