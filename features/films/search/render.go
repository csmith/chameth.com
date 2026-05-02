package search

import (
	"bytes"
	"embed"
	"html/template"

	"chameth.com/chameth.com/features/shortcodes"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("filmsearch.html.gotpl").ParseFS(templates, "filmsearch.html.gotpl"))

func RenderFromText(_ []string, _ *shortcodes.Context) (string, error) {
	return renderTemplate()
}

func renderTemplate() (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, nil)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
