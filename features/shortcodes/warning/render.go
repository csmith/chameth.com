package warning

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

var tmpl = template.Must(template.New("warning.html.gotpl").ParseFS(templates, "warning.html.gotpl"))

func RenderFromText(args []string, _ *shortcodes.Context) (string, error) {
	if len(args) < 1 {
		return "", errors.New("warning requires at least 1 argument (content)")
	}

	return renderTemplate(
		Data{
			Content: markdown.Render(args[0]),
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
