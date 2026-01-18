package rating

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strconv"

	"chameth.com/chameth.com/cmd/serve/content/shortcodes/context"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("rating.html.gotpl").ParseFS(templates, "rating.html.gotpl"))

func RenderFromText(args []string, _ *context.Context) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("rating requires at least 1 argument (value)")
	}

	numericRating, err := strconv.Atoi(args[0])
	if err != nil {
		return "", fmt.Errorf("invalid rating %s: %w", args[0], err)
	}

	return renderTemplate(Data{
		FilledStars: numericRating / 2,
		HalfStar:    numericRating%2 == 1,
		EmptyStars:  5 - ((numericRating + 1) / 2),
	})
}

func Render(rating int) (string, error) {
	return renderTemplate(Data{
		FilledStars: rating / 2,
		HalfStar:    rating%2 == 1,
		EmptyStars:  5 - ((rating + 1) / 2),
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
