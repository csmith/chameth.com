package templates

import (
	"bytes"
	"html/template"
)

var videoTemplate = template.Must(
	template.
		New("video.html.gotpl").
		ParseFS(
			templates,
			"video.html.gotpl",
		),
)

type VideoData struct {
	Src         string
	Description string
}

func RenderVideo(data VideoData) (string, error) {
	buf := &bytes.Buffer{}
	err := videoTemplate.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
