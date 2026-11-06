package figure

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"

	"chameth.com/chameth.com/features/media"
	"chameth.com/chameth.com/features/shortcodes"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("figure.html.gotpl").ParseFS(templates, "figure.html.gotpl"))

func RenderFromText(args []string, ctx *shortcodes.Context) (string, error) {
	if len(args) < 2 {
		return "", errors.New("figure requires at least 2 arguments (class, description)")
	}

	class := args[0]
	description := args[1]

	matchingMedia := ctx.MediaWithDescription(description)

	if len(matchingMedia) == 0 {
		return "", fmt.Errorf("figure media not found for description: %s", description)
	}

	picture, err := media.SelectPicture(matchingMedia, description)
	if err != nil {
		return "", err
	}

	return renderTemplate(
		Data{
			Class:       class,
			Sources:     picture.Sources,
			Src:         picture.Src,
			Description: description,
			Caption:     picture.Caption,
			Width:       picture.Width,
			Height:      picture.Height,
		},
	)
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
