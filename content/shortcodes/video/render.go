package video

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"chameth.com/chameth.com/content/shortcodes/common"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("video.html.gotpl").ParseFS(templates, "video.html.gotpl"))

func RenderFromText(args []string, ctx *common.Context) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("video requires at least 1 argument (description)")
	}

	description := args[0]

	mediaRelation := ctx.MediaWithDescription(description)
	if len(mediaRelation) != 1 {
		return "", fmt.Errorf("incorrect number of video files found for description %s (expected 1, got %d)", description, len(mediaRelation))
	}

	return renderTemplate(Data{
		Src:         mediaRelation[0].Path,
		Description: description,
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
