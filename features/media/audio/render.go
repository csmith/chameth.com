package audio

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"chameth.com/chameth.com/content/shortcodes/common"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("audio.html.gotpl").ParseFS(templates, "audio.html.gotpl"))

func RenderFromText(args []string, ctx *common.Context) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("audio requires at least 1 argument (description)")
	}

	description := args[0]

	mediaRelation := ctx.MediaWithDescription(description)
	if len(mediaRelation) != 1 {
		return "", fmt.Errorf("incorrect number of audio files found for description %s (expected 1, got %d)", description, len(mediaRelation))
	}

	caption := description
	if mediaRelation[0].Caption != nil && *mediaRelation[0].Caption != "" {
		caption = *mediaRelation[0].Caption
	}

	return renderTemplate(Data{
		Src:         mediaRelation[0].Path,
		Description: description,
		Caption:     caption,
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
