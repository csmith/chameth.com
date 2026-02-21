package update

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"chameth.com/chameth.com/content/markdown"
	"chameth.com/chameth.com/content/shortcodes/common"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("update.html.gotpl").ParseFS(templates, "update.html.gotpl"))

func RenderFromText(args []string, _ *common.Context) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("update requires at least 2 arguments (date, content)")
	}

	date := args[0]
	content := args[1]

	md, err := markdown.Render(content)
	if err != nil {
		return "", fmt.Errorf("failed to render update markdown: %w", err)
	}

	return renderTemplate(Data{
		Date:    date,
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
