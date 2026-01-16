package shortcodes

import (
	"fmt"
	"regexp"
	"strings"

	"chameth.com/chameth.com/cmd/serve/content/markdown"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/templates"
	"chameth.com/chameth.com/cmd/serve/db"
)

var (
	figureRegexp = regexp.MustCompile(`\{%\s*figure "(.*?)" "(.*?)"\s*%}`)
)

func renderFigure(input string, ctx *Context) (string, error) {
	res := input
	figures := figureRegexp.FindAllStringSubmatch(input, -1)
	for _, figure := range figures {
		class := strings.ReplaceAll(figure[1], `\\"`, `"`)
		description := strings.ReplaceAll(figure[2], `\\"`, `"`)

		matchingMedia := ctx.MediaWithDescription(description)

		if len(matchingMedia) == 0 {
			return "", fmt.Errorf("figure media not found for description: %s", description)
		}

		// Find the primary image (jpeg or png) and build sources
		var primaryMedia *db.MediaRelationWithDetails
		var sources []templates.FigureSource

		for i := range matchingMedia {
			m := &matchingMedia[i]
			switch m.ContentType {
			case "image/jpeg", "image/png":
				if primaryMedia == nil {
					primaryMedia = m
				}
			case "image/avif", "image/webp":
				sources = append(sources, templates.FigureSource{
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

		replacement, err := templates.RenderFigure(templates.FigureData{
			Class:       class,
			Sources:     sources,
			Src:         primaryMedia.Path,
			Description: description,
			Caption:     renderedCaption,
			Width:       *primaryMedia.Width,
			Height:      *primaryMedia.Height,
		})
		if err != nil {
			return "", fmt.Errorf("failed to render figure template: %w", err)
		}

		res = strings.Replace(res, figure[0], replacement, 1)
	}
	return res, nil
}
