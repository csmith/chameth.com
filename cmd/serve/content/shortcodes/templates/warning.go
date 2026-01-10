package templates

import (
	"bytes"
	"html/template"
)

var warningTemplate = template.Must(
	template.
		New("warning.html.gotpl").
		ParseFS(
			templates,
			"warning.html.gotpl",
		),
)

type WarningData struct {
	Content template.HTML
}

func RenderWarning(data WarningData) (string, error) {
	buf := &bytes.Buffer{}
	err := warningTemplate.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
