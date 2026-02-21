package filmsearch

import (
	"bytes"
	"embed"
	"html/template"

	"chameth.com/chameth.com/content/shortcodes/common"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("filmsearch.html.gotpl").ParseFS(templates, "filmsearch.html.gotpl"))

func RenderFromText(_ []string, _ *common.Context) (string, error) {
	return renderTemplate()
}

func renderTemplate() (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, nil)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
