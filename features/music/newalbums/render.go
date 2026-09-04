package newalbums

import (
	"bytes"
	"embed"
	"html/template"

	"chameth.com/chameth.com/features/shortcodes"
)

//go:embed newalbums.html.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("newalbums.html.gotpl").ParseFS(templates, "newalbums.html.gotpl"))

func render(_ []string, albums []Album, _ *shortcodes.Context) (string, error) {
	if len(albums) == 0 {
		return "", nil
	}

	return renderTemplate(Data{Albums: albums})
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
