package shortcodes

import (
	"fmt"

	"chameth.com/chameth.com/cmd/serve/content/markdown"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/templates"
	"chameth.com/chameth.com/cmd/serve/db"
)

func renderFigure(args []string, ctx *Context) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("figure requires at least 2 arguments (class, description)")
	}

	class := args[0]
	description := args[1]

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

	return replacement, nil
}
