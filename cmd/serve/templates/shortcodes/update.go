package shortcodes

import (
	"bytes"
	"html/template"
)

var updateTemplate = template.Must(
	template.
		New("update.html.gotpl").
		ParseFS(
			templates,
			"update.html.gotpl",
		),
)

type UpdateData struct {
	Date    string
	Content template.HTML
}

func RenderUpdate(data UpdateData) (string, error) {
	buf := &bytes.Buffer{}
	err := updateTemplate.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
