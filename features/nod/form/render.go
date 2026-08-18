package shortcode

import (
	"bytes"
	"embed"
	"html/template"

	"chameth.com/chameth.com/features/shortcodes"
	parenttemplates "chameth.com/chameth.com/templates"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("nod.html.gotpl").ParseFS(templates, "nod.html.gotpl"))

func RenderFromText(_ []string, ctx *shortcodes.Context) (string, error) {
	return renderTemplate(Data{
		Page: parenttemplates.SiteURL() + ctx.URL,
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
