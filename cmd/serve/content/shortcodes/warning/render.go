package warning

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

var tmpl = template.Must(template.New("warning.html.gotpl").ParseFS(templates, "warning.html.gotpl"))

func RenderFromText(args []string, _ *common.Context) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("warning requires at least 1 argument (content)")
	}

	content := args[0]

	md, err := markdown.Render(content)
	if err != nil {
		return "", fmt.Errorf("failed to render warning markdown: %w", err)
	}

	return renderTemplate(Data{
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
