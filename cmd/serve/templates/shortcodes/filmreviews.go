package shortcodes

import (
	"bytes"
	"html/template"
)

var filmReviewsTemplate = template.Must(
	template.New("filmreviews.html.gotpl").
		ParseFS(
			templates,
			"filmreviews.html.gotpl",
		),
)

type FilmReviewsData struct {
	Reviews []template.HTML
}

func RenderFilmReviews(data FilmReviewsData) (string, error) {
	buf := &bytes.Buffer{}
	err := filmReviewsTemplate.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
