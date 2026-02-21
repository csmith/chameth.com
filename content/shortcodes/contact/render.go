package contact

import (
	"bytes"
	"embed"
	"html/template"

	"chameth.com/chameth.com/content/shortcodes/common"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("contact.html.gotpl").ParseFS(templates, "contact.html.gotpl"))

func RenderFromText(_ []string, ctx *common.Context) (string, error) {
	return renderTemplate(Data{
		Page: ctx.URL,
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
