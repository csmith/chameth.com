package rating

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"math/rand"
	"strconv"

	"chameth.com/chameth.com/content/shortcodes/common"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("rating.html.gotpl").Funcs(template.FuncMap{
	"mod": func(a, b int) int { return a % b },
}).ParseFS(templates, "rating.html.gotpl"))

func RenderFromText(args []string, _ *common.Context) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("rating requires at least 1 argument (value)")
	}

	numericRating, err := strconv.Atoi(args[0])
	if err != nil {
		return "", fmt.Errorf("invalid rating %s: %w", args[0], err)
	}

	return Render(numericRating)
}

func Render(rating int) (string, error) {
	filledStars := rating / 2
	rotations := make([]int, filledStars)
	for i := range rotations {
		rotations[i] = rand.Intn(15)
	}

	return renderTemplate(Data{
		FilledStars: rotations,
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
