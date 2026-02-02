package figure

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"chameth.com/chameth.com/cmd/serve/content/markdown"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/common"
	"chameth.com/chameth.com/cmd/serve/db"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("figure.html.gotpl").ParseFS(templates, "figure.html.gotpl"))

func RenderFromText(args []string, ctx *common.Context) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("figure requires at least 2 arguments (class, description)")
	}

	class := args[0]
	description := args[1]

	matchingMedia := ctx.MediaWithDescription(description)

	if len(matchingMedia) == 0 {
		return "", fmt.Errorf("figure media not found for description: %s", description)
	}

	var primaryMedia *db.MediaRelationWithDetails
	var sources []Source

	for i := range matchingMedia {
		m := &matchingMedia[i]
		switch m.ContentType {
		case "image/jpeg", "image/png":
			if primaryMedia == nil {
				primaryMedia = m
			}
		case "image/avif", "image/webp":
			sources = append(sources, Source{
				Src:  m.Path,
				Type: m.ContentType,
			})
		}
	}

	if primaryMedia == nil {
		return "", fmt.Errorf("no jpeg or png image found for description: %s", description)
	}

	if primaryMedia.Width == nil || primaryMedia.Height == nil {
		return "", fmt.Errorf("image dimensions not set for description: %s", description)
	}

	caption := description
	if primaryMedia.Caption != nil && *primaryMedia.Caption != "" {
		caption = *primaryMedia.Caption
	}

	renderedCaption, err := markdown.Render(caption)
	if err != nil {
		return "", fmt.Errorf("failed to render figure caption markdown: %w", err)
	}

	return renderTemplate(Data{
		Class:       class,
		Sources:     sources,
		Src:         primaryMedia.Path,
		Description: description,
		Caption:     renderedCaption,
		Width:       *primaryMedia.Width,
		Height:      *primaryMedia.Height,
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
