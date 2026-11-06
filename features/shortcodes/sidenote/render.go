package sidenote

import (
	"bytes"
	"embed"
	"errors"
	"html/template"

	"chameth.com/chameth.com/content/markdown"
	"chameth.com/chameth.com/features/shortcodes"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("sidenote.html.gotpl").ParseFS(templates, "sidenote.html.gotpl"))

func RenderFromText(args []string, _ *shortcodes.Context) (string, error) {
	if len(args) < 2 {
		return "", errors.New("sidenote requires at least 2 arguments (title, content)")
	}

	return renderTemplate(
		Data{
			Title:   args[0],
			Content: markdown.Render(args[1]),
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
