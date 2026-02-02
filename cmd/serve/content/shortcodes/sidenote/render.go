package sidenote

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"chameth.com/chameth.com/cmd/serve/content/markdown"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/common"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("sidenote.html.gotpl").ParseFS(templates, "sidenote.html.gotpl"))

func RenderFromText(args []string, _ *common.Context) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("sidenote requires at least 2 arguments (title, content)")
	}

	title := args[0]
	content := args[1]

	md, err := markdown.Render(content)
	if err != nil {
		return "", fmt.Errorf("failed to render sidenote markdown: %w", err)
	}

	return renderTemplate(Data{
		Title:   title,
		Content: md,
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
