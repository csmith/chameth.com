package reviews

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"chameth.com/chameth.com/features/films"
	"chameth.com/chameth.com/features/films/review"
	"chameth.com/chameth.com/features/shortcodes"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("filmreviews.html.gotpl").ParseFS(templates, "filmreviews.html.gotpl"))

func render(_ []string, entries []films.ReviewEntry, _ *shortcodes.Context) (string, error) {
	var htmlList []template.HTML
	for _, entry := range entries {
		rendered, err := review.RenderEntry(entry)
		if err != nil {
			return "", fmt.Errorf("failed to render film review: %w", err)
		}

		htmlList = append(htmlList, template.HTML(rendered))
	}

	return renderTemplate(
		Data{
			Reviews: htmlList,
		},
	)
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
