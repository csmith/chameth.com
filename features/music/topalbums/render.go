package topalbums

import (
	"bytes"
	"embed"
	"html/template"

	"chameth.com/chameth.com/features/shortcodes"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("topalbums.html.gotpl").ParseFS(templates, "topalbums.html.gotpl"))

func render(_ []string, albums []Album, _ *shortcodes.Context) (string, error) {
	return renderTemplate(Data{Albums: albums})
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
