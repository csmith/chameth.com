package played

import (
	"bytes"
	"embed"
	"html/template"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("playedbgs.html.gotpl").ParseFS(templates, "playedbgs.html.gotpl"))

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
