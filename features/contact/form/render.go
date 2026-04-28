package form

import (
	"bytes"
	"embed"
	"html/template"

	"chameth.com/chameth.com/content/shortcodes/common"
	"chameth.com/chameth.com/features/contact"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("contact.html.gotpl").ParseFS(templates, "contact.html.gotpl"))

func RenderFromText(args []string, ctx *common.Context) (string, error) {
	preamble := ""
	if len(args) > 0 {
		preamble = args[0]
	}

	return renderTemplate(Data{
		Page:      ctx.URL,
		Preamble:  preamble,
		Timestamp: contact.SignedTimestamp(),
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
