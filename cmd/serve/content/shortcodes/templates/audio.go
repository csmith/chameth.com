package templates

import (
	"bytes"
	"html/template"
)

var audioTemplate = template.Must(
	template.
		New("audio.html.gotpl").
		ParseFS(
			templates,
			"audio.html.gotpl",
		),
)

type AudioData struct {
	Src         string
	Description string
	Caption     string
}

func RenderAudio(data AudioData) (string, error) {
	buf := &bytes.Buffer{}
	err := audioTemplate.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
