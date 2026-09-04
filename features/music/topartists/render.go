package topartists

import (
	"bytes"
	"embed"
	"html/template"

	"chameth.com/chameth.com/features/shortcodes"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("topartists.html.gotpl").ParseFS(templates, "topartists.html.gotpl"))

func render(_ []string, artists []Artist, _ *shortcodes.Context) (string, error) {
	return renderTemplate(Data{Artists: artists})
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
