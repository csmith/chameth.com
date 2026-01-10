package templates

import (
	"bytes"
	"html/template"
)

var sideNoteTemplate = template.Must(
	template.
		New("sidenote.html.gotpl").
		ParseFS(
			templates,
			"sidenote.html.gotpl",
		),
)

type SideNoteData struct {
	Title   string
	Content template.HTML
}

func RenderSideNote(data SideNoteData) (string, error) {
	buf := &bytes.Buffer{}
	err := sideNoteTemplate.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
