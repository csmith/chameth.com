package update

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

var tmpl = template.Must(template.New("update.html.gotpl").ParseFS(templates, "update.html.gotpl"))

func RenderFromText(args []string, _ *shortcodes.Context) (string, error) {
	if len(args) < 2 {
		return "", errors.New("update requires at least 2 arguments (date, content)")
	}

	return renderTemplate(
		Data{
			Date:    args[0],
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
